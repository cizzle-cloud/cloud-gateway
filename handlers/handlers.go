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

func ProxyRequestHandler(c *gin.Context, target string, fixedPath string) {
	targetURL, err := url.Parse(target)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid proxy target"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	log.Printf("[ PROXY ] Request received at %s at %s\n", c.Request.URL, time.Now())
	log.Printf("[ PROXY ] Target URL: %s", targetURL)
	log.Printf("[ PROXY ] Target URL Path: %s", targetURL.Path)
	log.Printf("[ PROXY ] C param Path: %s", c.Param("path"))
	log.Printf("[ PROXY ] Target URL host: %s", targetURL.Host)
	log.Printf("[ PROXY ] Target URL Scheme: %s", targetURL.Scheme)

	var newPath string
	if fixedPath != "" {
		newPath = targetURL.Path + c.Param("path") + fixedPath
	} else {
		newPath = targetURL.Path + c.Param("path")
	}

	c.Request.Header.Set("X-Forwarded-Host", c.Request.Host)

	c.Request.URL.Path = newPath
	c.Request.Host = targetURL.Host
	c.Request.URL.Host = targetURL.Host
	c.Request.URL.Scheme = targetURL.Scheme

	log.Printf("[ PROXY ] Request Path: %s", c.Request.URL.Path)
	log.Printf("[ PROXY ] X-Forwarded-Host: %s", c.Request.Host)
	log.Printf("[ PROXY ] Forwarding request to %s at %s\n", c.Request.URL, time.Now())

	log.Printf("[ PROXY ] X-Forwarded-Host: %s", c.Request.Header.Get("X-Forwarded-Host"))

	proxy.ServeHTTP(c.Writer, c.Request)
}

func RedirectHandler(c *gin.Context, url string) {
	c.Redirect(http.StatusFound, url)
}

func forwardRequest(c *gin.Context, target string) {
	targetURL, err := url.Parse(target)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid proxy target"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	log.Printf("[ PROXY ] X-Forwarded-Host: %s", c.Request.Host)

	log.Println("[ DOMAIN PROXY ] before")
	log.Printf("[ DOMAIN PROXY ] Host %s", c.Request.Host)
	log.Printf("[ DOMAIN PROXY ] URL host %s", c.Request.URL.Host)
	log.Printf("[ DOMAIN PROXY ] Scheme %s", c.Request.URL.Scheme)

	c.Request.Header.Set("X-Forwarded-Host", c.Request.Host)

	c.Request.Host = targetURL.Host
	c.Request.URL.Host = targetURL.Host
	c.Request.URL.Scheme = targetURL.Scheme

	log.Printf("[ PROXY ] X-Forwarded-Host: %s", c.Request.Host)

	log.Println("[ DOMAIN PROXY ] after")
	log.Printf("[ DOMAIN PROXY ] Host %s", c.Request.Host)
	log.Printf("[ DOMAIN PROXY ] URL host %s", c.Request.URL.Host)
	log.Printf("[ DOMAIN PROXY ] Scheme %s", c.Request.URL.Scheme)

	log.Printf("[ DOMAIN PROXY ] X-Forwarded-Host: %s", c.Request.Header.Get("X-Forwarded-Host"))
	proxy.ServeHTTP(c.Writer, c.Request)
}

func DomainProxyHandler(c *gin.Context, routes []route.DomainRoute) {
	host := c.Request.Host
	targetDomain := strings.Split(c.Request.Host, ":")[0]
	// TODO: Apply also middlware
	for _, r := range routes {
		if r.Domain == targetDomain {
			log.Printf("[ DOMAIN PROXY ] Host → %s, Domain → %s", host, r.Domain)
			forwardRequest(c, r.ProxyTarget)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "no backend found for domain"})
}
