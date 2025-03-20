package middleware

import (
	"log"
	"net/http"

	ratelimiter "github.com/cizzle-cloud/rate-limiter"
	"github.com/gin-gonic/gin"
)

//TODO: For future not rate limit only based per client IP?

func NewRateLimitMiddleware(rl *ratelimiter.RateLimiter, algo ratelimiter.RateLimitAlgo) gin.HandlerFunc {

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		if !rl.Exists(clientIP) {
			rl.Add(clientIP, algo)
		}
		if !rl.Allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			log.Printf("rate limit exceeded for client %s:", clientIP)
			c.Abort()
			return
		}

		c.Next()
	}
}
