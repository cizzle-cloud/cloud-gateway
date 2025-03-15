package route

import (
	"api_gateway/config"
	"api_gateway/handlers"
	"api_gateway/middleware"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// TODO: Fix and see what will happen with args

var MiddlewareRegistry = map[string]gin.HandlerFunc{
	"auth_middleware":      middleware.NoCacheMiddleware(),
	"ratelimit_middleware": middleware.NoCacheMiddleware(),
	"nocache_middleware":   middleware.NoCacheMiddleware(),
}

type Route struct {
	Method     string
	Prefix     string
	Middleware []gin.HandlerFunc
	// optional fields
	ProxyTarget    string
	RedirectTarget string
	FixedPath      string
}

func NewRoute(method, prefix string, middleware []gin.HandlerFunc) Route {
	return Route{
		Method:     method,
		Prefix:     prefix,
		Middleware: middleware,
	}
}

func (r Route) WithProxy(proxyTarget string) Route {
	r.ProxyTarget = proxyTarget
	return r
}

func (r Route) WithRedirect(redirectTarget string) Route {
	r.RedirectTarget = redirectTarget
	return r
}

func (r Route) WithFixedPath(fixedPath string) Route {
	r.FixedPath = fixedPath
	return r
}

type DomainRoute struct {
	Domain      string
	ProxyTarget string
	Middleware  []gin.HandlerFunc
}

func NewDomainRoute(domain, proxyTarget string, middleware []gin.HandlerFunc) DomainRoute {
	return DomainRoute{
		Domain:      domain,
		ProxyTarget: proxyTarget,
		Middleware:  middleware,
	}
}

type MiddlewareGroup struct {
	Middleware []gin.HandlerFunc
}

type RouteRegistry struct {
	Routes       []Route
	DomainRoutes []DomainRoute
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
				if handler, ok := MiddlewareRegistry[mw]; ok {
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
		if handler, ok := MiddlewareRegistry[mw]; ok {
			resolvedMiddleware = append(resolvedMiddleware, handler)
		}
	}
	return resolvedMiddleware
}

func (rr *RouteRegistry) buildRoutes(cfg config.Config) {
	var routes []Route

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
			route := NewRoute(r.Method, r.Prefix+"/*path", resolvedMiddleware).WithProxy(r.ProxyTarget)
			routes = append(routes, route)
			continue
		}

		if r.RedirectTarget != "" {
			route := NewRoute(r.Method, r.Prefix+"/*path", resolvedMiddleware).WithRedirect(r.RedirectTarget)
			routes = append(routes, route)
			continue
		}

		// Register individual paths under the prefix
		for _, path := range r.Paths {

			if path.ProxyTarget != "" && path.RedirectTarget != "" {
				log.Fatal("[ERROR] Both Proxy and Redirect Target is defined and this is not allowed. Program will exit.")
			}

			fixedPath := path.Path
			route := NewRoute(path.Method, r.Prefix+fixedPath, resolvedMiddleware).
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
	var domainRoutes []DomainRoute

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
		domainRoute := NewDomainRoute(r.Domain, r.ProxyTarget, resolvedMiddleware)
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

func DomainProxyHandler(c *gin.Context, routes []DomainRoute) {
	host := c.Request.Host
	targetDomain := strings.Split(c.Request.Host, ":")[0]
	// TODO: Apply also middlware
	for _, r := range routes {
		if r.Domain == targetDomain {
			log.Printf("[ DOMAIN PROXY ] Host → %s, Domain → %s", host, r.Domain)
			handlers.ForwardRequest(c, r.ProxyTarget)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "no backend found for domain"})
}

func (rr *RouteRegistry) RegisterDomainRoutes(r *gin.Engine) {
	r.NoRoute(func(c *gin.Context) {
		DomainProxyHandler(c, rr.DomainRoutes)
	})
}

// for _, route := range rr.DomainRoutes {
// 	log.Println("[INFO] handlers route", route.Domain)
// 	handler := func(c *gin.Context) {
// 		handlers.DomainProxyHandler(c, route.Domain, route.ProxyTarget)
// 	}

// 	handlers := append(route.Middleware, handler)
// 	log.Println("[INFO] registering route", route.Domain)
// 	log.Println("[INFO] registering route", route.Domain, handlers)
// 	r.Use(handlers...)
