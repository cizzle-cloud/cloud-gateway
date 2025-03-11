package route

import (
	"api_gateway/config"
	"api_gateway/handlers"
	"api_gateway/middleware"
	"log"

	"github.com/gin-gonic/gin"
)

// TODO: Fix and see what will happen with args

var MiddlewareRegistry = map[string]gin.HandlerFunc{
	"auth_middleware":      middleware.NoCacheMiddleware(),
	"ratelimit_middleware": middleware.NoCacheMiddleware(),
	"nocache_middleware":   middleware.NoCacheMiddleware(),
}

type Route struct {
	Method      string
	Prefix      string
	Middleware  []gin.HandlerFunc
	ProxyTarget string
	// optional fields
	FixedPath string
}

// TODO: Investigate if we need pointer to Route or Route
// struct (memory in use vs escape to heap allocation)

func NewRoute(method, prefix string, middleware []gin.HandlerFunc, proxyTarget string) Route {
	return Route{
		Method:      method,
		Prefix:      prefix,
		Middleware:  middleware,
		ProxyTarget: proxyTarget,
	}
}

func (r Route) WithFixedPath(fixedPath string) Route {
	r.FixedPath = fixedPath
	return r
}

type MiddlewareGroup struct {
	Middleware []gin.HandlerFunc
}

type RouteRegistry struct {
	Routes []Route
}

func (rr *RouteRegistry) FromConfig(cfg *config.InputConfig) {
	var routes []Route

	for _, r := range cfg.Routes {
		var resolvedMiddleware []gin.HandlerFunc
		// TODO: Where would you put a sanity check for the config.json/yaml?
		// Resolve middleware (either direct list or from group)
		if r.MiddlewareGroup != "" {
			if groupMiddleware, exists := cfg.MiddlewareGroups[r.MiddlewareGroup]; exists {
				for _, mw := range groupMiddleware {
					if handler, ok := MiddlewareRegistry[mw]; ok {
						resolvedMiddleware = append(resolvedMiddleware, handler)
					}
				}
			}
		} else {
			for _, mw := range r.Middleware {
				if handler, ok := MiddlewareRegistry[mw]; ok {
					resolvedMiddleware = append(resolvedMiddleware, handler)
				}
			}
		}

		// If the entire prefix is a proxy route (no specific paths)
		if r.ProxyTarget != "" {
			route := NewRoute(r.Method, r.Prefix+"/*path", resolvedMiddleware, r.ProxyTarget)
			log.Println("ROUTE FIXED PATH", route.FixedPath)

			routes = append(routes, route)
			continue
		}

		// Register individual paths under the prefix
		for _, path := range r.Paths {
			fixedPath := path.Path
			route := NewRoute(path.Method, r.Prefix+fixedPath, resolvedMiddleware, path.ProxyTarget).
				WithFixedPath(fixedPath)
			log.Println("ROUTE", route.FixedPath)
			routes = append(routes, route)
		}
	}

	rr.Routes = routes
}

func (rr *RouteRegistry) RegisterRoutes(r *gin.Engine) {
	for _, route := range rr.Routes {
		handlers := append(route.Middleware, func(c *gin.Context) {
			handlers.ProxyRequestHandler(c, route.ProxyTarget, route.FixedPath)
		})
		r.Handle(route.Method, route.Prefix, handlers...)
	}
}
