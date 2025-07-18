package middleware

import (
	"bytes"
	"cloud_gateway/config"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func NewForwardAuthMiddleware(cfg *config.ForwardAuthConfig) gin.HandlerFunc {

	var client *http.Client
	if cfg.CertFilepath != "" {
		client = &http.Client{Transport: getTransport(cfg.CertFilepath)}
	} else {
		client = &http.Client{}
	}

	return func(c *gin.Context) {

		// Bypass auth for CORS preflight requests
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), cfg.Timeout)
		defer cancel()

		// Prepare body
		var body io.Reader
		if cfg.ForwardBody {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to read request body"})
				return
			}

			// c.Request.Body is now drained. We need to refill it for later use downstream
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			body = bytes.NewBuffer(bodyBytes)
		}

		// Prepare auth request
		authReq, err := http.NewRequestWithContext(ctx, cfg.Method, cfg.Url, body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to create auth request"})
			return
		}

		// Attach cookies to request from original request
		for _, cookieName := range cfg.AddCookiesToRequest {
			if cookie, err := c.Request.Cookie(cookieName); err == nil {
				authReq.AddCookie(cookie)
			}
		}

		// Attach request headers
		for _, h := range cfg.RequestHeaders {
			if val := c.GetHeader(h); val != "" {
				authReq.Header.Set(h, val)
			}
		}

		if cfg.TrustForwardHeader {
			authReq.Header.Set("X-Forwarded-Host", c.Request.Host)
			authReq.Header.Set("X-Forwarded-Method", c.Request.Method)
			authReq.Header.Set("X-Forwarded-Uri", c.Request.RequestURI)
			authReq.Header.Set("X-Forwarded-For", c.ClientIP())

			scheme := "http"
			if c.Request.TLS != nil {
				scheme = "https"
			}

			authReq.Header.Set("X-Forwarded-Proto", scheme)
		}

		// Send the request
		resp, err := client.Do(authReq)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": fmt.Sprintf("auth service unreachable: %v", err)})
			return
		}
		defer resp.Body.Close()

		// If not authorized, return the response as-is
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			// Propagate headers
			for _, h := range cfg.ResponseHeaders {
				if val := resp.Header.Get(h); val != "" {
					c.Header(h, val)
				}
			}

			// Propagate cookies
			for _, cookieName := range cfg.AddCookiesToResponse {
				for _, cookie := range resp.Cookies() {
					if cookie.Name == cookieName {
						http.SetCookie(c.Writer, cookie)
					}
				}
			}

			c.Status(resp.StatusCode)
			io.Copy(c.Writer, resp.Body)
			c.Abort()
			return
		}

		// Authorized — optionally copy some response headers
		for _, h := range cfg.ResponseHeaders {
			if val := resp.Header.Get(h); val != "" {
				c.Header(h, val)
			}
		}

		// Set cookies to response
		for _, cookieName := range cfg.AddCookiesToResponse {
			for _, cookie := range resp.Cookies() {
				if cookie.Name == cookieName {
					http.SetCookie(c.Writer, cookie)
				}
			}
		}

		c.Next()
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
