package router

import (
	"net/http"

	"github.com/cizzle-cloud/cloud-gateway/internal/middleware"
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

func (gr *GinRouter) Handle(method, path string, handler http.Handler, mws ...middleware.HTTPFunc) {
	finalHandler := middleware.Chain[middleware.HTTPHandler](handler, mws...)
	wrapped := func(c *gin.Context) {
		finalHandler.ServeHTTP(c.Writer, c.Request)
	}

	gr.engine.Handle(method, path, wrapped)
}

func (gr *GinRouter) NoRoute(handler http.Handler, mws ...middleware.HTTPFunc) {
	finalHandler := middleware.Chain[middleware.HTTPHandler](handler, mws...)
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
