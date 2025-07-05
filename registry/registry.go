package registry

import (
	"cloud_gateway/config"
	"cloud_gateway/handlers"
	"cloud_gateway/middleware"
	"cloud_gateway/route"
	"log"
	"path"

	ratelimiter "github.com/cizzle-cloud/rate-limiter"
	"github.com/gin-gonic/gin"
)

const (
	RouteHandle       = 0
	RouteInvalidRoute = 1
)

type RouteRegistry struct {
	Routes       []route.Route
	DomainRoutes []route.DomainRoute
}

func (rr *RouteRegistry) FromConfig(cfg *config.Config) {
	rr.ParseRoutes(cfg)
	rr.ParseDomainRoutes(cfg)
}

func resolveMiddlewareGroup(middlewareGroup string, cfg *config.Config) []gin.HandlerFunc {
	grp, ok := cfg.MiddlewareGroups[middlewareGroup]
	if !ok {
		return nil
	}

	return resolveMiddlewareList(*grp, cfg)
}

func resolveMiddleware(mw string, cfg *config.Config) gin.HandlerFunc {
	var handler gin.HandlerFunc

	if rateLimitCfg, ok := cfg.RateLimiters[mw]; ok {
		algo, rl := ParseRateLimitCfg(rateLimitCfg)
		handler = middleware.NewRateLimitMiddleware(algo, rl)
	} else if forwardAuthCfg, ok := cfg.ForwardAuth[mw]; ok {
		handler = middleware.NewForwardAuthMiddleware(forwardAuthCfg)
	} else {
		log.Fatalf("[ERROR] Unknown or unsupported middleware: %s", mw)
	}

	return handler
}

func resolveMiddlewareList(mwl []string, cfg *config.Config) []gin.HandlerFunc {
	var handlers []gin.HandlerFunc

	for _, mw := range mwl {
		handlers = append(handlers, resolveMiddleware(mw, cfg))
	}

	return handlers
}

func ParseRateLimitCfg(cfg *config.RateLimitConfig) (*ratelimiter.RateLimiter, ratelimiter.RateLimitAlgo) {
	var algo ratelimiter.RateLimitAlgo

	switch algoType := cfg.Algorithm; algoType {
	case "fixed_window_counter":
		algo = ratelimiter.NewFixedWindowCounter(cfg.Limit, cfg.WindowSize)
	case "token_bucket":
		algo = ratelimiter.NewTokenBucket(cfg.Capacity, cfg.RefillTokens, cfg.RefillInterval)
	}

	rl := ratelimiter.NewRateLimiter(cfg.Ttl, cfg.CleanupInterval)

	return rl, algo
}

func (rr *RouteRegistry) ParseRoutes(cfg *config.Config) {
	var routes []route.Route

	for _, r := range cfg.Routes {

		resolvedMiddleware := append(
			resolveMiddlewareGroup(r.MiddlewareGroup, cfg),
			resolveMiddlewareList(r.Middleware, cfg)...,
		)

		if r.ProxyTarget != "" {
			routes = append(routes, handleProxyRoute(r, resolvedMiddleware))
			continue
		}

		if r.RedirectTarget != "" {
			routes = append(
				routes,
				route.NewRoute(
					r.Method,
					r.Prefix,
					r.Prefix,
					resolvedMiddleware,
				).WithRedirect(r.RedirectTarget, r.RedirectCode),
			)
			continue
		}

		routes = append(routes, handlePathRoutes(r, cfg, resolvedMiddleware)...)
	}

	rr.Routes = routes
}

// Handle Proxy Target for prefix routes where no specific paths are defined
func handleProxyRoute(r *config.RouteConfig, resolvedMiddleware []gin.HandlerFunc) route.Route {
	if r.Prefix == "" || r.Prefix == "/" {
		return route.NewRoute(r.Method, r.Prefix, r.Prefix, resolvedMiddleware).WithProxy(r.ProxyTarget)
	}

	return route.NewRoute(r.Method, r.Prefix, r.Prefix+"/*path", resolvedMiddleware).WithProxy(r.ProxyTarget)
}

// Handle individual paths under the prefix
func handlePathRoutes(r *config.RouteConfig, cfg *config.Config, resolvedRouteMiddleware []gin.HandlerFunc) []route.Route {
	var pathRoutes []route.Route

	for _, path := range r.Paths {

		resolvedPathMiddleware := append(
			resolveMiddlewareGroup(path.MiddlewareGroup, cfg),
			resolveMiddlewareList(path.Middleware, cfg)...,
		)

		resolvedMiddleware := append(
			append([]gin.HandlerFunc{}, resolvedRouteMiddleware...),
			resolvedPathMiddleware...,
		)

		fixedPath := path.Path
		var pathRoute route.Route
		if path.ProxyTarget != "" {
			pathRoute = route.NewRoute(path.Method, r.Prefix, r.Prefix+fixedPath+"/*path", resolvedMiddleware).
				WithFixedPath(fixedPath).WithProxy(path.ProxyTarget)
		}

		if path.RedirectTarget != "" {
			pathRoute = route.NewRoute(
				path.Method,
				r.Prefix,
				r.Prefix+fixedPath,
				resolvedMiddleware,
			).WithFixedPath(fixedPath).WithRedirect(path.RedirectTarget, path.RedirectCode)
		}

		pathRoutes = append(pathRoutes, pathRoute)
	}

	return pathRoutes
}

func (rr *RouteRegistry) ParseDomainRoutes(cfg *config.Config) {
	var domainRoutes []route.DomainRoute

	for _, r := range cfg.DomainRoutes {
		resolvedMiddleware := append(
			resolveMiddlewareGroup(r.MiddlewareGroup, cfg),
			resolveMiddlewareList(r.Middleware, cfg)...,
		)

		domainPaths := make([]route.DomainPath, 0, len(r.Paths))
		for _, path := range r.Paths {
			resolvedPathMiddleware := resolveMiddlewareList(path.Middleware, cfg)
			domainPath := route.NewDomainPath(path.Path, path.Method, resolvedPathMiddleware)
			domainPaths = append(domainPaths, domainPath)
		}

		domainRoutes = append(
			domainRoutes,
			route.NewDomainRoute(r.Domain, r.ProxyTarget, resolvedMiddleware).WithPaths(domainPaths),
		)
	}

	rr.DomainRoutes = domainRoutes
}

func getRouteHandler(route route.Route) (gin.HandlerFunc, int8) {
	switch {
	case route.ProxyTarget != "":
		//TODO: I think evaluation inside path.Clean method is wrong
		return func(c *gin.Context) {
			handlers.ProxyRequestHandler(c, route.ProxyTarget, path.Clean(c.Param("path")+route.FixedPath))
		}, RouteHandle

	case route.RedirectTarget != "":
		return func(c *gin.Context) {
			handlers.RedirectHandler(c, route.RedirectTarget, route.RedirectCode)
		}, RouteHandle
	default:
		return nil, RouteInvalidRoute

	}
}

func (rr *RouteRegistry) RegisterRoutes(r *gin.Engine) {
	for _, route := range rr.Routes {
		handler, routeType := getRouteHandler(route)

		switch routeType {
		case RouteHandle:
			handlerFuncs := append(route.Middleware, handler)
			r.Handle(route.Method, route.RelativePath, handlerFuncs...)
		case RouteInvalidRoute:
			log.Fatal("[ERROR] Invalid/Unknown route configuration")
		}
	}
}

func (rr *RouteRegistry) RegisterDomainRoutes(r *gin.Engine) {
	if len(rr.DomainRoutes) == 0 {
		return
	}
	r.NoRoute(func(c *gin.Context) {
		handlers.DomainProxyHandler(c, rr.DomainRoutes)
	})
}
