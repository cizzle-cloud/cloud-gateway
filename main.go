package main

import (
	"api_gateway/config"
	"api_gateway/handlers"
	"api_gateway/middleware"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		err.Handle()
		return
	}

	r := gin.Default()
	r.GET(
		"/",
		func(c *gin.Context) {
			c.Redirect(http.StatusFound, cfg.AuthPagePath)
		})
	r.GET(
		fmt.Sprintf("/%s/*path", cfg.AuthPagePath),
		middleware.NoCacheMiddleware(),
		middleware.AuthMiddleware(cfg),
		func(c *gin.Context) {
			handlers.ProxyRequestHandler(c, cfg.AuthPageURL)
		})
	r.GET(
		fmt.Sprintf("/%s/*path", cfg.AuthPath),
		middleware.AuthMiddleware(cfg),
		func(c *gin.Context) {
			handlers.ProxyRequestHandler(c, cfg.AuthURL)
		})

	r.NoRoute(handlers.Handle404)

	r.Run()
}
