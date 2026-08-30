package router

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cizzle-cloud/cloud-gateway/internal/middleware"
)

type GinRouter struct {
	engine *gin.Engine
}

func NewGinRouter(ginMode string, trustedProxies []string) *GinRouter {
	gin.SetMode(ginMode)
	engine := gin.Default()
	if err := engine.SetTrustedProxies(trustedProxies); err != nil {
		panic(fmt.Sprintf("invalid trusted proxies: %v", err))
	}
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

func (gr *GinRouter) Run(addr string) error {
	return gr.engine.Run(addr)
}

func (gr *GinRouter) RunTLS(addr, certFile, keyFile string) error {
	return gr.engine.RunTLS(addr, certFile, keyFile)
}
