package handlers

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

func ProxyRequestHandler(c *gin.Context, target string, fixedPath string) {
	targetURL, err := url.Parse(target)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid proxy target"})
		return
	}

	log.Printf("[ PROXY ] Target URL: %s", targetURL)

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

	c.Request.URL.Path = newPath
	c.Request.Host = targetURL.Host
	c.Request.URL.Host = targetURL.Host
	c.Request.URL.Scheme = targetURL.Scheme
	c.Request.Header.Set("X-Forwarded-Host", c.Request.Header.Get("Host"))
	log.Printf("[ PROXY ] Request Path: %s", c.Request.URL.Path)
	log.Printf("[ PROXY ] X-Forwarded-Host: %s", c.Request.Header.Get("Host"))

	log.Printf("[ PROXY ] Forwarding request to %s at %s\n", c.Request.URL, time.Now())
	proxy.ServeHTTP(c.Writer, c.Request)
}

func Handle404(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"message": "Page Not Found"})
}
