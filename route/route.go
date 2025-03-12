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

type MiddlewareGroup struct {
	Middleware []gin.HandlerFunc
}

type RouteRegistry struct {
	Routes []Route
}

func (rr *RouteRegistry) FromConfig(cfg config.Config) {
	var routes []Route

	for _, r := range cfg.Routes {

		if r.ProxyTarget != "" && r.RedirectTarget != "" {
			log.Fatal("[ERROR] Both Proxy and Redirect Target is defined and this is not allowed. Route will be skipped.")
			continue
		}

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
				log.Fatal("[ERROR] Both Proxy and Redirect Target is defined and this is not allowed. Route will be skipped.")
				continue
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
