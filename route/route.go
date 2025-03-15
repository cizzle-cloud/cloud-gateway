package route

import (
	"api_gateway/middleware"

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
