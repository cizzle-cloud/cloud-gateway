package middleware

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/cizzle-cloud/cloud-gateway/internal/config"
	"github.com/cizzle-cloud/cloud-gateway/internal/request"
	"github.com/cizzle-cloud/cloud-gateway/internal/response"
)

func NewForwardAuthMiddleware(cfg *config.ForwardAuthConfig) HTTPFunc {

	var client *http.Client
	if cfg.CertFilepath != "" {
		client = &http.Client{Transport: getTransport(cfg.CertFilepath)}
	} else {
		client = &http.Client{}
	}

	return func(next HTTPHandler) HTTPHandler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Bypass auth for CORS preflight requests
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// Create context with timeout
			ctx, cancel := context.WithTimeout(r.Context(), cfg.Timeout)
			defer cancel()

			// Prepare body
			var body io.Reader
			if cfg.ForwardBody {
				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					response.WriteJSONError(w, http.StatusInternalServerError, "failed to read request body")
					return
				}

				// c.Request.Body is now drained. We need to refill it for later use downstream
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				body = bytes.NewBuffer(bodyBytes)
			}

			// Prepare auth request
			authReq, err := http.NewRequestWithContext(ctx, cfg.Method, cfg.Url, body)
			if err != nil {
				response.WriteJSONError(w, http.StatusInternalServerError, "failed to create auth request")
				return
			}

			// Attach cookies to request from original request
			for _, cookieName := range cfg.AddCookiesToRequest {
				if cookie, err := r.Cookie(cookieName); err == nil {
					authReq.AddCookie(cookie)
				}
			}

			// Attach headers from original request
			for _, h := range cfg.RequestHeaders {
				if val := r.Header.Get(h); val != "" {
					authReq.Header.Set(h, val)
				}
			}

			if cfg.TrustForwardHeader {
				authReq.Header.Set("X-Forwarded-Host", r.Host)
				authReq.Header.Set("X-Forwarded-Method", r.Method)
				authReq.Header.Set("X-Forwarded-Uri", r.RequestURI)
				authReq.Header.Set("X-Forwarded-For", request.ClientIP(r))

				scheme := "http"
				if r.TLS != nil {
					scheme = "https"
				}

				authReq.Header.Set("X-Forwarded-Proto", scheme)
			}

			// Send the request
			resp, err := client.Do(authReq)
			if err != nil {
				response.WriteJSONError(w, http.StatusServiceUnavailable, fmt.Sprintf("auth service unreachable: %v", err))
				return
			}
			defer resp.Body.Close()

			// If not authorized, return the response as-is
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				// Propagate headers
				for _, h := range cfg.ResponseHeaders {
					if val := resp.Header.Get(h); val != "" {
						w.Header().Set(h, val)
					}
				}

				// Propagate cookies
				for _, cookieName := range cfg.AddCookiesToResponse {
					for _, cookie := range resp.Cookies() {
						if cookie.Name == cookieName {
							http.SetCookie(w, cookie)
						}
					}
				}

				w.WriteHeader(resp.StatusCode)
				io.Copy(w, resp.Body)
				return
			}

			// Authorized — optionally copy some response headers
			for _, h := range cfg.ResponseHeaders {
				if val := resp.Header.Get(h); val != "" {
					w.Header().Set(h, val)
				}
			}

			// Set cookies to response
			for _, cookieName := range cfg.AddCookiesToResponse {
				for _, cookie := range resp.Cookies() {
					if cookie.Name == cookieName {
						http.SetCookie(w, cookie)
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getTransport(certFilepath string) *http.Transport {

	certData, err := os.ReadFile(certFilepath)
	if err != nil {
		log.Fatalf("[FORWARD AUTH] Failed to read cert file: %v", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(certData)

	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	return &http.Transport{
		TLSClientConfig: tlsConfig,
	}
}
