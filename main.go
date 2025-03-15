package main

import (
	"api_gateway/config"
	"api_gateway/route"
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		err.Handle()
		return
	}

	r := gin.Default()
	rr := &route.RouteRegistry{}
	rr.FromConfig(cfg)
	rr.RegisterRoutes(r)
	rr.RegisterDomainRoutes(r)
	r.Run(fmt.Sprintf("127.0.0.1:%s", cfg.Env.Port))
}
