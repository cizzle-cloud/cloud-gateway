package route

import (
	"github.com/gin-gonic/gin"
)

type Route struct {
	Method       string
	Prefix       string
	RelativePath string
	Middleware   []gin.HandlerFunc
	// optional fields
	ProxyTarget    string
	RedirectTarget string
	RedirectCode   int
	FixedPath      string
}

func NewRoute(method, prefix, relativePath string, middleware []gin.HandlerFunc) Route {
	return Route{
		Method:       method,
		Prefix:       prefix,
		RelativePath: relativePath,
		Middleware:   middleware,
	}
}

func (r Route) WithProxy(proxyTarget string) Route {
	r.ProxyTarget = proxyTarget
	return r
}

func (r Route) WithRedirect(redirectTarget string, redirectCode int) Route {
	r.RedirectTarget = redirectTarget
	r.RedirectCode = redirectCode
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
