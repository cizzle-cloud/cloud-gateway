package route

import "github.com/cizzle-cloud/cloud-gateway/internal/middleware"

type Route struct {
	Method       string
	Prefix       string
	RelativePath string
	Middleware   []middleware.HTTPFunc
	// optional fields
	ProxyTarget    string
	RedirectTarget string
	RedirectCode   int
	FixedPath      string
}

func NewRoute(method, prefix, relativePath string, mws []middleware.HTTPFunc) *Route {
	return &Route{
		Method:       method,
		Prefix:       prefix,
		RelativePath: relativePath,
		Middleware:   mws,
	}
}

func (r *Route) WithProxy(proxyTarget string) *Route {
	r.ProxyTarget = proxyTarget
	return r
}

func (r *Route) WithRedirect(redirectTarget string, redirectCode int) *Route {
	r.RedirectTarget = redirectTarget
	r.RedirectCode = redirectCode
	return r
}

func (r *Route) WithFixedPath(fixedPath string) *Route {
	r.FixedPath = fixedPath
	return r
}

type DomainPath struct {
	Path       string
	Method     string
	Middleware []middleware.HTTPFunc
}

func NewDomainPath(path, method string, mws []middleware.HTTPFunc) DomainPath {
	return DomainPath{
		Path:       path,
		Method:     method,
		Middleware: mws,
	}
}

type DomainRoute struct {
	Domain      string
	ProxyTarget string
	Middleware  []middleware.HTTPFunc
	// optional fields
	Paths []DomainPath
}

func NewDomainRoute(domain, proxyTarget string, mws []middleware.HTTPFunc) *DomainRoute {
	return &DomainRoute{
		Domain:      domain,
		ProxyTarget: proxyTarget,
		Middleware:  mws,
	}
}

func (dr *DomainRoute) WithPaths(paths []DomainPath) *DomainRoute {
	dr.Paths = paths
	return dr
}
