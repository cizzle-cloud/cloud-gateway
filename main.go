package main

import (
	"api_gateway/config"
	"api_gateway/registry"
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	env, err := config.LoadEnv()
	if err != nil {
		err.Handle()
		return
	}

	cfg, err := config.LoadConfig(env.ConfigFilepath, env.ConfigFileType)
	if err != nil {
		err.Handle()
		return
	}

	r := gin.Default()
	rr := &registry.RouteRegistry{}
	rr.FromConfig(cfg)
	rr.RegisterRoutes(r)
	rr.RegisterDomainRoutes(r)
	r.Run(fmt.Sprintf("%s:%s", cfg.Env.Host, cfg.Env.Port))
}
