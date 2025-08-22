package middleware

import (
	"log"
	"net"
	"net/http"

	"github.com/cizzle-cloud/cloud-gateway/internal/router"
	ratelimiter "github.com/cizzle-cloud/rate-limiter"
)

//TODO: For future not rate limit only based per client IP?

func NewRateLimitMiddleware(rl *ratelimiter.RateLimiter, algo ratelimiter.RateLimitAlgo) router.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				clientIP = r.RemoteAddr
			}

			if !rl.Exists(clientIP) {
				rl.Add(clientIP, algo)
			}
			if !rl.Allow(clientIP) {
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				log.Printf("[MIDDLEWARE] rate limit exceeded for client %s", clientIP)
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
