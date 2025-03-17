package registry

import (
	"api_gateway/config"
	"api_gateway/handlers"
	"api_gateway/route"
	"log"

	"github.com/gin-gonic/gin"
)

type RouteRegistry struct {
	Routes       []route.Route
	DomainRoutes []route.DomainRoute
}

func (rr *RouteRegistry) FromConfig(cfg config.Config) {
	rr.buildRoutes(cfg)
	rr.buildDomainRoutes(cfg)
}

func resolveMiddlewareGroup(middlewareGroup string, middlewareGroups map[string]config.ConfigMiddlewareGroup) []gin.HandlerFunc {
	var resolvedMiddleware []gin.HandlerFunc
	if middlewareGroup != "" {
		if mws, exists := middlewareGroups[middlewareGroup]; exists {
			for _, mw := range mws {
				if handler, ok := route.MiddlewareRegistry[mw]; ok {
					resolvedMiddleware = append(resolvedMiddleware, handler)
				}
			}
		}
	}
	return resolvedMiddleware
}

func resolveMiddleware(middleware []string) []gin.HandlerFunc {
	var resolvedMiddleware []gin.HandlerFunc
	for _, mw := range middleware {
		if handler, ok := route.MiddlewareRegistry[mw]; ok {
			resolvedMiddleware = append(resolvedMiddleware, handler)
		}
	}
	return resolvedMiddleware
}

func (rr *RouteRegistry) buildRoutes(cfg config.Config) {
	var routes []route.Route

	for _, r := range cfg.Routes {

		if r.ProxyTarget != "" && r.RedirectTarget != "" {
			log.Fatal("[ERROR] Both Proxy and Redirect Target is defined and this is not allowed. Program will exit.")
		}

		var resolvedMiddleware []gin.HandlerFunc
		if r.MiddlewareGroup != "" {
			resolvedMiddleware = resolveMiddlewareGroup(r.MiddlewareGroup, cfg.MiddlewareGroups)

		} else {
			resolvedMiddleware = resolveMiddleware(r.Middleware)
		}

		log.Println("[INFO] Resolved middleware", r.Prefix, resolvedMiddleware)

		// If the entire prefix is a proxy route (no specific paths)
		if r.ProxyTarget != "" {
			route := route.NewRoute(r.Method, r.Prefix+"/*path", resolvedMiddleware).WithProxy(r.ProxyTarget)
			routes = append(routes, route)
			continue
		}

		if r.RedirectTarget != "" {
			route := route.NewRoute(r.Method, r.Prefix+"/*path", resolvedMiddleware).WithRedirect(r.RedirectTarget)
			routes = append(routes, route)
			continue
		}

		// Register individual paths under the prefix
		for _, path := range r.Paths {

			if path.ProxyTarget != "" && path.RedirectTarget != "" {
				log.Fatal("[ERROR] Both Proxy and Redirect Target is defined and this is not allowed. Program will exit.")
			}

			fixedPath := path.Path
			route := route.NewRoute(path.Method, r.Prefix+fixedPath, resolvedMiddleware).
				WithFixedPath(fixedPath)

			if path.ProxyTarget != "" {
				route = route.WithProxy(path.ProxyTarget)
			}

			if path.RedirectTarget != "" {
				route = route.WithRedirect(path.RedirectTarget)
			}

			routes = append(routes, route)
		}
	}

	rr.Routes = routes
}

func (rr *RouteRegistry) buildDomainRoutes(cfg config.Config) {
	var domainRoutes []route.DomainRoute

	for _, r := range cfg.DomainRoutes {
		if r.ProxyTarget == "" || r.Domain == "" {
			log.Fatal("[ERROR] Domain or ProxyTarget is missing. Program will exit.")
		}

		var resolvedMiddleware []gin.HandlerFunc
		if r.MiddlewareGroup != "" {
			resolvedMiddleware = resolveMiddlewareGroup(r.MiddlewareGroup, cfg.MiddlewareGroups)

		} else {
			resolvedMiddleware = resolveMiddleware(r.Middleware)
		}
		log.Println("[INFO] Resolved middleware", r.Domain, resolvedMiddleware)
		domainRoute := route.NewDomainRoute(r.Domain, r.ProxyTarget, resolvedMiddleware)
		domainRoutes = append(domainRoutes, domainRoute)
	}

	rr.DomainRoutes = domainRoutes
}

func (rr *RouteRegistry) RegisterRoutes(r *gin.Engine) {
	for _, route := range rr.Routes {
		var handler gin.HandlerFunc
		if route.ProxyTarget != "" {
			handler = func(c *gin.Context) {
				handlers.ProxyRequestHandler(c, route.ProxyTarget, route.FixedPath)
			}
		} else if route.RedirectTarget != "" {
			handler = func(c *gin.Context) {
				handlers.RedirectHandler(c, route.RedirectTarget)
			}
		}
		handlers := append(route.Middleware, handler)
		r.Handle(route.Method, route.Prefix, handlers...)
	}
}

func (rr *RouteRegistry) RegisterDomainRoutes(r *gin.Engine) {
	r.NoRoute(func(c *gin.Context) {
		handlers.DomainProxyHandler(c, rr.DomainRoutes)
	})
}
