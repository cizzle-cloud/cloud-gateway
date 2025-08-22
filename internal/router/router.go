package router

import "net/http"

type Router interface {
}

type MiddlewareFunc func(http.Handler) http.Handler

type HTTPRouter interface {
	Router
	Run(addr string)
	RunTLS(addr, certFile, keyFile string)
	Handle(method, path string, handler http.Handler, middleware ...MiddlewareFunc)
	NoRoute(handler http.Handler, middleware ...MiddlewareFunc)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type WebSocketRouter interface {
	Router
}

type GRPCRouter interface {
	Router
}
