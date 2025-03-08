package route

import (
	"api_gateway/handlers"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Route struct {
	Method     string
	Prefix     string
	Middleware []gin.HandlerFunc
	// optional fields
	ProxyTarget string
}

// TODO: Investigate if we need pointer to Route or Route
// struct (memory in use vs escape to heap allocation)

func NewRoute(method, prefix string, middleware []gin.HandlerFunc) Route {
	return Route{
		Method:     method,
		Prefix:     prefix,
		Middleware: middleware,
	}
}

func (r Route) WithProxy(target string) Route {
	r.ProxyTarget = target
	return r
}

type MiddlewareGroup struct {
	Middleware []gin.HandlerFunc
}

type RouteRegistry struct {
	Routes []Route
}

func (rr *RouteRegistry) RegisterRoutes(r *gin.Engine) {
	for _, route := range rr.Routes {
		handlers := append(route.Middleware, routeHandler(route))
		r.Handle(route.Method, route.Prefix, handlers...)
	}
}

func routeHandler(route Route) gin.HandlerFunc {
	if route.ProxyTarget != "" {
		return func(c *gin.Context) {
			handlers.ProxyRequestHandler(c, route.ProxyTarget)
		}
	}

	return func(c *gin.Context) {
		c.JSON(
			http.StatusOK,
			gin.H{"message": fmt.Sprintf("handling reguest for : %s", route.Prefix)},
		)
	}
}
