package middleware

import (
	"context"
	"net/http"

	"golang.org/x/net/websocket"
)

type Handler interface {
}

type HTTPFunc = func(HTTPHandler) HTTPHandler
type WSFunc = func(WSHandler) WSHandler
type GRPCFunc = func(GRPCHandler) GRPCHandler

type HTTPHandler interface {
	Handler
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type WSHandler interface {
	Handler
	ServeWS(conn *websocket.Conn)
}

type GRPCHandler interface {
	Handler
	ServeGRPC(ctx context.Context, req interface{}) (interface{}, error)
}

// Chain applies middleware in the correct order (last middleware wraps first)
func Chain[T Handler](handler T, middleware ...func(T) T) T {
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	return handler
}
