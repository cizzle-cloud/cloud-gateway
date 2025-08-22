package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type GinRouter struct {
	engine *gin.Engine
}

func NewGinRouter(ginMode string, trustedProxies []string) *GinRouter {
	gin.SetMode(ginMode)
	engine := gin.Default()
	engine.SetTrustedProxies(trustedProxies)
	return &GinRouter{engine: engine}
}

func (gr *GinRouter) Handle(method, path string, handler http.Handler, middleware ...MiddlewareFunc) {
	finalHandler := applyMiddleware(handler, middleware...)
	wrapped := func(c *gin.Context) {
		finalHandler.ServeHTTP(c.Writer, c.Request)
	}

	gr.engine.Handle(method, path, wrapped)
}

func (gr *GinRouter) NoRoute(handler http.Handler, middleware ...MiddlewareFunc) {
	finalHandler := applyMiddleware(handler, middleware...)
	wrapped := func(c *gin.Context) {
		finalHandler.ServeHTTP(c.Writer, c.Request)
	}

	gr.engine.NoRoute(wrapped)
}

func (gr *GinRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gr.engine.ServeHTTP(w, r)
}

func (gr *GinRouter) Run(addr string) {
	gr.engine.Run(addr)
}

func (gr *GinRouter) RunTLS(addr, certFile, keyFile string) {
	gr.engine.RunTLS(addr, certFile, keyFile)
}

// applyMiddleware applies middleware in the correct order (last middleware wraps first)
func applyMiddleware(handler http.Handler, middleware ...MiddlewareFunc) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	return handler
}
