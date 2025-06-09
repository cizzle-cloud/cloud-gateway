package handlers

import (
	"api_gateway/route"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func ProxyRequestHandler(c *gin.Context, target, targetPath string) {
	targetURL, err := url.Parse(target)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid proxy target"})
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
		req.Header.Set("X-Forwarded-Host", c.Request.Host)

		log.Printf("[PROXY] Forwarding request to %s at %s\n", req.URL, time.Now())
		log.Printf("[PROXY] X-Forwarded-Host: %s", req.Header.Get("X-Forwarded-Host"))
	}

	log.Printf("[PROXY] Request received at %s at %s\n", c.Request.URL, time.Now())
	log.Printf("[PROXY] Target URL: %s", targetURL)

	proxy.ServeHTTP(c.Writer, c.Request)
}

func RedirectHandler(c *gin.Context, url string, code int) {
	c.Redirect(code, url)
}

func DomainProxyHandler(c *gin.Context, routes []route.DomainRoute) {
	targetDomain := strings.Split(c.Request.Host, ":")[0]
	reqPath := c.Request.URL.Path
	reqMethod := c.Request.Method

	for _, r := range routes {
		if r.Domain != targetDomain {
			continue
		}

		// Apply middlware
		for _, mw := range r.Middleware {
			mw(c)
			if c.IsAborted() {
				return
			}
		}

		for _, p := range r.Paths {
			if p.Path != reqPath || p.Method != reqMethod {
				continue
			}
			for _, pmw := range p.Middleware {
				pmw(c)
				if c.IsAborted() {
					return
				}
			}
		}

		ProxyRequestHandler(c, r.ProxyTarget, reqPath)
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "no backend found for domain"})
}

func BaseProxyRequestHandler(c *gin.Context, target string) {
	targetURL, err := url.Parse(target)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid proxy target"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	proxy.Director = func(req *http.Request) {
		// Modify request parameters
		req.Host = targetURL.Host
		req.URL.Host = targetURL.Host
		req.URL.Scheme = targetURL.Scheme

		// Forward original host
		req.Header.Set("X-Forwarded-Host", c.Request.Host)

		log.Printf("[PROXY] Forwarding request to %s at %s\n", req.URL, time.Now())
		log.Printf("[PROXY] X-Forwarded-Host: %s", req.Header.Get("X-Forwarded-Host"))
	}

	log.Printf("[PROXY] Request received at %s at %s\n", c.Request.URL, time.Now())
	log.Printf("[PROXY] Target URL: %s", targetURL)

	proxy.ServeHTTP(c.Writer, c.Request)
}
