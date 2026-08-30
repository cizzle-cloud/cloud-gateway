package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	ratelimiter "github.com/cizzle-cloud/rate-limiter"

	"github.com/cizzle-cloud/cloud-gateway/internal/config"
	"github.com/cizzle-cloud/cloud-gateway/internal/request"
	"github.com/cizzle-cloud/cloud-gateway/internal/response"
)

// TODO: For future not rate limit only based per client IP?
func NewRateLimitMiddleware(c *request.Context, cfg *config.RateLimitConfig) HTTPFunc {
	rl := ratelimiter.NewRateLimiter(cfg.TTL, cfg.CleanupInterval)
	return func(next HTTPHandler) HTTPHandler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := request.ClientIP(c, r)
			record := rl.GetOrCreate(clientIP, createRateLimitAlgorithm(cfg))
			if !record.Allow() {
				log.Printf("[MIDDLEWARE] rate limit exceeded for client %q", sanitizeLogValue(clientIP))
				response.WriteJSONError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			// continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// sanitizeLogValue returns a quoted, escaped form of value suitable for safe logging.
func sanitizeLogValue(value string) string {
	return fmt.Sprintf("%q", strings.ReplaceAll(strings.ReplaceAll(value, "\r", "\\r"), "\n", "\\n"))
}

func createRateLimitAlgorithm(cfg *config.RateLimitConfig) ratelimiter.RateLimitAlgo {
	var algo ratelimiter.RateLimitAlgo

	switch algoType := cfg.Algorithm; algoType {
	case "fixed_window_counter":
		algo = ratelimiter.NewFixedWindowCounter(cfg.Limit, cfg.WindowSize)
	case "token_bucket":
		algo = ratelimiter.NewTokenBucket(cfg.Capacity, cfg.RefillTokens, cfg.RefillInterval)
	}
	return algo
}
