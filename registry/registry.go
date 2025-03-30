package registry

import (
	"api_gateway/config"
	"api_gateway/handlers"
	"api_gateway/middleware"
	"api_gateway/route"
	"log"

	ratelimiter "github.com/cizzle-cloud/rate-limiter"
	"github.com/gin-gonic/gin"
)

const (
	RouteHandle       = 0
	RouteNoRoute      = 1
	RouteInvalidRoute = 2
)

type RouteRegistry struct {
	Routes       []route.Route
	DomainRoutes []route.DomainRoute
}

func (rr *RouteRegistry) FromConfig(cfg config.Config) {
	rr.ParseRoutes(cfg)
	rr.ParseDomainRoutes(cfg)
}

func resolveMiddlewareGroup(middlewareGroup string, middlewareGroups map[string]config.MiddlewareGroupConfig) []gin.HandlerFunc {
	var resolvedMiddleware []gin.HandlerFunc
	if middlewareGroup != "" {
		if mwg, exists := middlewareGroups[middlewareGroup]; exists {
			for _, mw := range mwg {
				log.Println(mw)
				// resolvedMiddleware = append(resolvedMiddleware, handler)
			}
		}
	}
	return resolvedMiddleware
}

func ParseRateLimitCfg(cfg config.RateLimitConfig) (*ratelimiter.RateLimiter, ratelimiter.RateLimitAlgo) {
	var algo ratelimiter.RateLimitAlgo

	switch algoType := cfg.Algorithm; algoType {
	case "fixed_window_counter":
		algo = ratelimiter.NewFixedWindowCounter(cfg.Limit, cfg.WindowSize)
	case "token_bucket":
		algo = ratelimiter.NewTokenBucket(cfg.Capacity, cfg.RefillTokens, cfg.RefillInterval)
	default:
		log.Fatalf("[ERROR] Unknown / Unsupported rate limit algorithm: %s.", algo)
	}

	rl := ratelimiter.NewRateLimiter(cfg.Ttl, cfg.CleanupInterval)

	return rl, algo
}

func resolveMiddleware(mws []string, cfg config.Config) []gin.HandlerFunc {
	var resolvedMiddleware []gin.HandlerFunc
	for _, mw := range mws {
		var handler gin.HandlerFunc

		if rateLimitCfg, ok := cfg.RateLimiters[mw]; ok {
			algo, rl := ParseRateLimitCfg(rateLimitCfg)
			handler = middleware.NewRateLimitMiddleware(algo, rl)
		} else if authCfg, ok := cfg.Auth[mw]; ok {
			handler = middleware.NewAuthMiddleware(authCfg)
		} else if noCachePolicyCfg, ok := cfg.NoCachePolicies[mw]; ok {
			handler = middleware.NewNoCacheMiddleware(noCachePolicyCfg)
		} else {
			log.Fatalf("[ERROR] Unknown or unsupported middleware: %s", mw)
		}

		resolvedMiddleware = append(resolvedMiddleware, handler)
	}
	return resolvedMiddleware
}

func (rr *RouteRegistry) ParseRoutes(cfg config.Config) {
	var routes []route.Route

	for _, r := range cfg.Routes {

		if r.ProxyTarget != "" && r.RedirectTarget != "" {
			log.Fatal("[ERROR] Both Proxy and Redirect Target is defined and this is not allowed. Program will exit.")
		}

		var resolvedMiddleware []gin.HandlerFunc
		if r.MiddlewareGroup != "" {
			resolvedMiddleware = resolveMiddlewareGroup(r.MiddlewareGroup, cfg.MiddlewareGroups)

		} else {
			resolvedMiddleware = resolveMiddleware(r.Middleware, cfg)
		}

		log.Println("[INFO] Resolved middleware", r.Prefix, resolvedMiddleware)

		// If the entire prefix is a proxy route (no specific paths)
		if r.ProxyTarget != "" {
			if r.Prefix == "" || r.Prefix == "/" {
				if r.Method != "" {
					log.Fatal("[ERROR] Base route with proxy target has method defined.")
				} else {
					route := route.NewRoute(r.Method, r.Prefix, r.Prefix, resolvedMiddleware).WithProxy(r.ProxyTarget)
					routes = append(routes, route)
					continue
				}
			}

			route := route.NewRoute(r.Method, r.Prefix, r.Prefix+"/*path", resolvedMiddleware).WithProxy(r.ProxyTarget)
			routes = append(routes, route)
			continue
		}

		if r.RedirectTarget != "" {
			route := route.NewRoute(r.Method, r.Prefix, r.Prefix, resolvedMiddleware).WithRedirect(r.RedirectTarget)
			routes = append(routes, route)
			continue
		}

		// Register individual paths under the prefix
		for _, path := range r.Paths {

			if path.ProxyTarget != "" && path.RedirectTarget != "" {
				log.Fatal("[ERROR] Both Proxy and Redirect Target is defined and this is not allowed. Program will exit.")
			}

			fixedPath := path.Path
			var pathRoute route.Route
			if path.ProxyTarget != "" {
				pathRoute = route.NewRoute(path.Method, r.Prefix, r.Prefix+fixedPath+"/*path", resolvedMiddleware).
					WithFixedPath(fixedPath).WithProxy(path.ProxyTarget)
			}

			if path.RedirectTarget != "" {
				pathRoute = route.NewRoute(path.Method, r.Prefix, r.Prefix+fixedPath, resolvedMiddleware).
					WithFixedPath(fixedPath).WithRedirect(path.RedirectTarget)
			}

			routes = append(routes, pathRoute)
		}
	}

	rr.Routes = routes
}

func (rr *RouteRegistry) ParseDomainRoutes(cfg config.Config) {
	var domainRoutes []route.DomainRoute

	for _, r := range cfg.DomainRoutes {
		if r.ProxyTarget == "" || r.Domain == "" {
			log.Fatal("[ERROR] Domain or ProxyTarget is missing. Program will exit.")
		}

		var resolvedMiddleware []gin.HandlerFunc
		if r.MiddlewareGroup != "" {
			resolvedMiddleware = resolveMiddlewareGroup(r.MiddlewareGroup, cfg.MiddlewareGroups)

		} else {
			resolvedMiddleware = resolveMiddleware(r.Middleware, cfg)
		}
		log.Println("[INFO] Resolved middleware", r.Domain, resolvedMiddleware)
		domainRoute := route.NewDomainRoute(r.Domain, r.ProxyTarget, resolvedMiddleware)
		domainRoutes = append(domainRoutes, domainRoute)
	}

	rr.DomainRoutes = domainRoutes
}

func getRouteHandler(route route.Route) (gin.HandlerFunc, int8) {
	switch {
	case route.Prefix == "" || route.Prefix == "/":
		return func(c *gin.Context) {
			handlers.BaseRouteProxyHandler(c, route.ProxyTarget)
		}, RouteNoRoute

	case route.ProxyTarget != "":
		return func(c *gin.Context) {
			handlers.ProxyRequestHandler(c, route.ProxyTarget, route.FixedPath)
		}, RouteHandle

	case route.RedirectTarget != "":
		return func(c *gin.Context) {
			handlers.RedirectHandler(c, route.RedirectTarget)
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
		case RouteNoRoute:
			r.NoRoute(handler)
		case RouteInvalidRoute:
			// TODO Understand fatal vs panic
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
