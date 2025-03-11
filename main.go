package main

import (
	"api_gateway/config"
	"api_gateway/route"

	"github.com/gin-gonic/gin"
)

func main() {

	inputCfg, err := config.LoadInputConfig("config_template.yaml", "yaml")
	if err != nil {
		err.Handle()
		return
	}

	r := gin.Default()
	rr := &route.RouteRegistry{}
	rr.FromConfig(inputCfg)
	rr.RegisterRoutes(r)
	r.Run()
}
