package router

import (
	"net/http"

	"github.com/cizzle-cloud/cloud-gateway/internal/middleware"
)

type Router interface{}

type HTTPRouter interface {
	Router
	Run(addr string) error
	RunTLS(addr, certFile, keyFile string) error
	Handle(method, path string, handler http.Handler, middleware ...middleware.HTTPFunc)
	NoRoute(handler http.Handler, middleware ...middleware.HTTPFunc)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type WebSocketRouter interface {
	Router
}

type GRPCRouter interface {
	Router
}
