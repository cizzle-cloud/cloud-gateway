package main

import (
	"fmt"

	"github.com/cizzle-cloud/cloud-gateway/internal/config"
	"github.com/cizzle-cloud/cloud-gateway/internal/registry"
	"github.com/cizzle-cloud/cloud-gateway/internal/router"
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

	httpRouter := router.NewGinRouter(cfg.Env.Mode, cfg.Env.TrustedProxies)

	rr := &registry.RouteRegistry{}
	rr.FromConfig(cfg)
	rr.RegisterRoutes(httpRouter)
	rr.RegisterDomainRoutes(httpRouter)
	addr := fmt.Sprintf("%s:%v", cfg.Env.Host, cfg.Env.Port)
	certFilepath := cfg.Env.CertFilepath
	keyFilepath := cfg.Env.KeyFilepath
	if certFilepath == "" || keyFilepath == "" {
		httpRouter.Run(addr)
	} else {

		httpRouter.RunTLS(addr, certFilepath, keyFilepath)
	}

}
