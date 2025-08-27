package middleware

import (
	"log"
	"net/http"

	"github.com/cizzle-cloud/cloud-gateway/internal/request"
	"github.com/cizzle-cloud/cloud-gateway/internal/response"
	ratelimiter "github.com/cizzle-cloud/rate-limiter"
)

//TODO: For future not rate limit only based per client IP?

func NewRateLimitMiddleware(c *request.Context, rl *ratelimiter.RateLimiter, algo ratelimiter.RateLimitAlgo) HTTPFunc {
	return func(next HTTPHandler) HTTPHandler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := request.ClientIP(c, r)

			if !rl.Exists(clientIP) {
				rl.Add(clientIP, algo)
			}
			if !rl.Allow(clientIP) {
				log.Printf("[MIDDLEWARE] rate limit exceeded for client %s", clientIP)
				response.WriteJSONError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			// continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// return func(c *gin.Context) {
// clientIP := c.ClientIP()
// if !rl.Exists(clientIP) {
// rl.Add(clientIP, algo)
// }
// if !rl.Allow(clientIP) {
// c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
// log.Printf("[MIDDLEWARE] rate limit exceeded for client %s:", clientIP)
// c.Abort()
// return
// }
//
// c.Next()
// }
