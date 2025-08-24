package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/cizzle-cloud/cloud-gateway/internal/middleware"
	"github.com/cizzle-cloud/cloud-gateway/internal/route"
)

func ProxyRequest(target, targetPath string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetURL, err := url.Parse(target)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "invalid proxy target")
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		proxy.Director = func(req *http.Request) {
			// Modify request parameters
			req.URL.Path = targetURL.Path + targetPath
			req.Host = targetURL.Host
			req.URL.Host = targetURL.Host
			req.URL.Scheme = targetURL.Scheme

			// Forward original host
			req.Header.Set("X-Forwarded-Host", r.Host)

			log.Printf("[PROXY] Forwarding request to %s at %s\n", req.URL, time.Now())
			log.Printf("[PROXY] X-Forwarded-Host: %s", req.Header.Get("X-Forwarded-Host"))
		}

		log.Printf("[PROXY] Request received at %s at %s\n", r.URL, time.Now())
		log.Printf("[PROXY] Target URL: %s", targetURL)

		proxy.ServeHTTP(w, r)
	})
}

// Redirect creates an HTTP handler that redirects to the specified url with the given status code
func Redirect(url string, code int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, code)
	})
}

// ProxyDomain creates an HTTP handler that routes requests based on  domain configuration
func ProxyDomain(routes []route.DomainRoute) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetDomain := strings.Split(r.Host, ":")[0]
		reqPath := r.URL.Path
		reqMethod := r.Method

		for _, route := range routes {
			if route.Domain != targetDomain {
				continue
			}

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Find matching path and request method
				for _, p := range route.Paths {
					if p.Path != reqPath || p.Method != reqMethod {
						continue
					}

					// Apply path-level middleware and proxy
					finalHandler := middleware.Chain[middleware.HTTPHandler](
						ProxyRequest(route.ProxyTarget, reqPath),
						p.Middleware...,
					)
					finalHandler.ServeHTTP(w, r)
					return
				}

				// If this line reached, path matched domain but no specific path route
				// Proxy to the target then with the original path
				ProxyRequest(route.ProxyTarget, reqPath).ServeHTTP(w, r)
			})

			// Apply domain-level middleware
			finalHandler := middleware.Chain[middleware.HTTPHandler](handler, route.Middleware...)
			finalHandler.ServeHTTP(w, r)
			return
		}

		writeJSONError(w, http.StatusNotFound, "no backend found for domain")
	})

}

// Helper function to write JSON error responses
func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := map[string]string{"error": message}
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		log.Printf("[ERROR] Failed to encode JSON error response: %v", err)
	}
}
