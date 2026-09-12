package handler

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/cizzle-cloud/cloud-gateway/internal/middleware"
	"github.com/cizzle-cloud/cloud-gateway/internal/response"
	"github.com/cizzle-cloud/cloud-gateway/internal/route"
)

func ProxyRequest(target, targetPath string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetURL, err := url.Parse(target)
		if err != nil {
			response.WriteJSONError(w, http.StatusInternalServerError, "invalid proxy target")
			return
		}

		proxy := &httputil.ReverseProxy{
			Rewrite: func(pr *httputil.ProxyRequest) {
				req := pr.Out

				// Modify request parameters
				req.URL.Path = targetURL.Path + targetPath
				req.Host = targetURL.Host
				req.URL.Host = targetURL.Host
				req.URL.Scheme = targetURL.Scheme

				// Forward original host
				req.Header.Set("X-Forwarded-Host", pr.In.Host)

				log.Printf("[PROXY] Forwarding request to %q at %s\n", sanitizeLogValue(req.URL.String()), time.Now())
				log.Printf("[PROXY] X-Forwarded-Host: %q", sanitizeLogValue(req.Header.Get("X-Forwarded-Host")))
			},
		}

		log.Printf("[PROXY] Request received at %q at %s\n", sanitizeLogValue(r.URL.String()), time.Now())
		log.Printf("[PROXY] Target URL: %q", sanitizeLogValue(targetURL.String()))

		proxy.ServeHTTP(w, r)
	})
}

// Redirect creates an HTTP handler that redirects to the specified URL with the given status code
func Redirect(redirectURL string, code int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, redirectURL, code)
	})
}

// ProxyDomain creates an HTTP handler that routes requests based on  domain configuration
func ProxyDomain(routes []*route.DomainRoute) http.Handler {
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
		response.WriteJSONError(w, http.StatusNotFound, "no backend found for domain")
	})
}

// sanitizeLogValue returns a quoted, escaped form of value suitable for safe logging.
func sanitizeLogValue(value string) string {
	return fmt.Sprintf("%q", strings.ReplaceAll(strings.ReplaceAll(value, "\r", "\\r"), "\n", "\\n"))
}
