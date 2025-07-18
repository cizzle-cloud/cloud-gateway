package main

import (
	"cloud_gateway/config"
	"cloud_gateway/registry"
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

	gin.SetMode(cfg.Env.GinMode)
	r := gin.Default()
	r.SetTrustedProxies(cfg.Env.TrustedProxies)
	rr := &registry.RouteRegistry{}
	rr.FromConfig(cfg)
	rr.RegisterRoutes(r)
	rr.RegisterDomainRoutes(r)

	addr := fmt.Sprintf("%s:%v", cfg.Env.Host, cfg.Env.Port)
	certFilepath := cfg.Env.CertFilepath
	keyFilepath := cfg.Env.KeyFilepath
	if certFilepath == "" || keyFilepath == "" {
		r.Run(addr)
	} else {

		r.RunTLS(addr, certFilepath, keyFilepath)
	}

}
