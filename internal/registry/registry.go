package registry

import (
	"log"
	"net/http"
	"path"

	"github.com/cizzle-cloud/cloud-gateway/internal/config"
	"github.com/cizzle-cloud/cloud-gateway/internal/handler"
	"github.com/cizzle-cloud/cloud-gateway/internal/middleware"
	"github.com/cizzle-cloud/cloud-gateway/internal/request"
	"github.com/cizzle-cloud/cloud-gateway/internal/route"
	"github.com/cizzle-cloud/cloud-gateway/internal/router"
)

const (
	RouteHandle       = 0
	RouteInvalidRoute = 1
)

type RouteRegistry struct {
	Context      *request.Context
	Routes       []*route.Route
	DomainRoutes []*route.DomainRoute
}

func New(cfg *config.Config) *RouteRegistry {
	// TODO: validate early in config and remove error returning from NewContext
	c, _ := request.NewContext(cfg.Env.TrustedProxies, cfg.Env.TrustHeaders)
	routes := parseRoutes(c, cfg)
	domainRoutes := parseDomainRoutes(c, cfg)
	return &RouteRegistry{
		Context:      c,
		Routes:       routes,
		DomainRoutes: domainRoutes,
	}
}

func resolveMiddlewareGroup(middlewareGroup string, c *request.Context, cfg *config.Config) []middleware.HTTPFunc {
	grp, ok := cfg.MiddlewareGroups[middlewareGroup]
	if !ok {
		return nil
	}

	return resolveMiddlewareList(*grp, c, cfg)
}

func resolveMiddleware(mw string, c *request.Context, cfg *config.Config) middleware.HTTPFunc {
	var h middleware.HTTPFunc

	if rateLimitCfg, ok := cfg.RateLimiters[mw]; ok {
		h = middleware.NewRateLimitMiddleware(c, rateLimitCfg)
	} else if forwardAuthCfg, ok := cfg.ForwardAuth[mw]; ok {
		h = middleware.NewForwardAuthMiddleware(c, forwardAuthCfg)
	} else {
		log.Fatalf("[ERROR] Unknown or unsupported middleware: %s", mw)
	}

	return h
}

func resolveMiddlewareList(mwl []string, c *request.Context, cfg *config.Config) []middleware.HTTPFunc {
	var handlers []middleware.HTTPFunc

	for _, mw := range mwl {
		handlers = append(handlers, resolveMiddleware(mw, c, cfg))
	}

	return handlers
}

func parseRoutes(c *request.Context, cfg *config.Config) []*route.Route {
	var routes []*route.Route

	for _, r := range cfg.Routes {
		resolvedMiddleware := append(
			resolveMiddlewareGroup(r.MiddlewareGroup, c, cfg),
			resolveMiddlewareList(r.Middleware, c, cfg)...,
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

		routes = append(routes, handlePathRoutes(c, r, cfg, resolvedMiddleware)...)
	}

	return routes
}

// Handle Proxy Target for prefix routes where no specific paths are defined
func handleProxyRoute(r *config.RouteConfig, resolvedMiddleware []middleware.HTTPFunc) *route.Route {
	if r.Prefix == "" || r.Prefix == "/" {
		return route.NewRoute(r.Method, r.Prefix, r.Prefix, resolvedMiddleware).WithProxy(r.ProxyTarget)
	}

	return route.NewRoute(r.Method, r.Prefix, r.Prefix+"/*path", resolvedMiddleware).WithProxy(r.ProxyTarget)
}

// Handle individual paths under the prefix
func handlePathRoutes(c *request.Context, r *config.RouteConfig, cfg *config.Config, resolvedRouteMiddleware []middleware.HTTPFunc) []*route.Route {
	var pathRoutes []*route.Route

	for _, path := range r.Paths {
		resolvedPathMiddleware := append(
			resolveMiddlewareGroup(path.MiddlewareGroup, c, cfg),
			resolveMiddlewareList(path.Middleware, c, cfg)...,
		)

		resolvedMiddleware := append(
			append([]middleware.HTTPFunc{}, resolvedRouteMiddleware...),
			resolvedPathMiddleware...,
		)

		fixedPath := path.Path
		var pathRoute *route.Route
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

func parseDomainRoutes(c *request.Context, cfg *config.Config) []*route.DomainRoute {
	var domainRoutes []*route.DomainRoute

	for _, r := range cfg.DomainRoutes {
		resolvedMiddleware := append(
			resolveMiddlewareGroup(r.MiddlewareGroup, c, cfg),
			resolveMiddlewareList(r.Middleware, c, cfg)...,
		)

		domainPaths := make([]route.DomainPath, 0, len(r.Paths))
		for _, path := range r.Paths {
			resolvedPathMiddleware := resolveMiddlewareList(path.Middleware, c, cfg)
			domainPath := route.NewDomainPath(path.Path, path.Method, resolvedPathMiddleware)
			domainPaths = append(domainPaths, domainPath)
		}

		domainRoutes = append(
			domainRoutes,
			route.NewDomainRoute(r.Domain, r.ProxyTarget, resolvedMiddleware).WithPaths(domainPaths),
		)
	}

	return domainRoutes
}

func getRouteHandler(r *route.Route) (h http.Handler, routeType int8) {
	if r == nil {
		return nil, RouteInvalidRoute
	}

	switch {
	case r.ProxyTarget != "":
		// TODO: implement the dynamic part of the path route.FixedPath+c.Param("path") coming from the request
		return handler.ProxyRequest(r.ProxyTarget, path.Clean(r.FixedPath)), RouteHandle

	case r.RedirectTarget != "":
		return handler.Redirect(r.RedirectTarget, r.RedirectCode), RouteHandle
	default:
		return nil, RouteInvalidRoute
	}
}

func (rr *RouteRegistry) RegisterRoutes(rtr router.HTTPRouter) {
	for _, route := range rr.Routes {
		h, routeType := getRouteHandler(route)

		switch routeType {
		case RouteHandle:
			rtr.Handle(route.Method, route.RelativePath, h, route.Middleware...)
		case RouteInvalidRoute:
			log.Fatal("[ERROR] Invalid/Unknown route configuration")
		}
	}
}

func (rr *RouteRegistry) RegisterDomainRoutes(rtr router.HTTPRouter) {
	if len(rr.DomainRoutes) == 0 {
		return
	}
	rtr.NoRoute(
		handler.ProxyDomain(rr.DomainRoutes),
	)
}
