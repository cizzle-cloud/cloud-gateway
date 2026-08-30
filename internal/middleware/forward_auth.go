package middleware

import (
	"bytes"
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

func NewForwardAuthMiddleware(c *request.Context, cfg *config.ForwardAuthConfig) HTTPFunc {
	client := newAuthClient(cfg)

	return func(next HTTPHandler) HTTPHandler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Bypass auth for CORS preflight requests
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			authReq, refilledBody, err := prepareForwardAuthRequest(c, cfg, r)
			if err != nil {
				response.WriteJSONError(w, http.StatusInternalServerError, err.Error())
				return
			}
			if refilledBody != nil {
				r.Body = refilledBody
			}

			// #nosec G704: auth endpoint URL comes from operator-supplied config, not user input
			resp, err := client.Do(authReq)
			if err != nil {
				response.WriteJSONError(w, http.StatusServiceUnavailable, fmt.Sprintf("auth service unreachable: %v", err))
				return
			}
			defer resp.Body.Close()

			propagateForwardAuthResponse(w, cfg, resp)

			// If not authorized, return the response as-is
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				w.WriteHeader(resp.StatusCode)
				_, _ = io.Copy(w, resp.Body)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func newAuthClient(cfg *config.ForwardAuthConfig) *http.Client {
	client := &http.Client{Timeout: cfg.Timeout}
	if cfg.CertFilepath != "" {
		client.Transport = getTransport(cfg.CertFilepath)
	}
	return client
}

func prepareForwardAuthRequest(c *request.Context, cfg *config.ForwardAuthConfig, r *http.Request) (*http.Request, io.ReadCloser, error) {
	var body io.Reader
	var refilledBody io.ReadCloser

	if cfg.ForwardBody {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read request body: %w", err)
		}

		// r.Body is now drained. Refill it for later use downstream
		refilledBody = io.NopCloser(bytes.NewBuffer(bodyBytes))
		body = bytes.NewBuffer(bodyBytes)
	}

	authReq, err := http.NewRequestWithContext(r.Context(), cfg.Method, cfg.URL, body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create auth request: %w", err)
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
		authReq.Header.Set("X-Forwarded-For", request.ClientIP(c, r))

		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}

		authReq.Header.Set("X-Forwarded-Proto", scheme)
	}

	return authReq, refilledBody, nil
}

func propagateForwardAuthResponse(w http.ResponseWriter, cfg *config.ForwardAuthConfig, resp *http.Response) {
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
