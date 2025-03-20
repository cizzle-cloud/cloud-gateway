package middleware

import (
	"api_gateway/config"
	"log"

	"github.com/gin-gonic/gin"
)

func NewNoCacheMiddleware(noCachePolicy config.NoCachePolicyConfig) gin.HandlerFunc {
	return func(c *gin.Context) { log.Println(noCachePolicy) }
}

// Middleware that disables cache of responses for protected endpoints.
func NoCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0, s-maxage=0")
		c.Writer.Header().Set("Pragma", "no-cache")
		c.Writer.Header().Set("Expires", "0")

		c.Next()
	}
}
