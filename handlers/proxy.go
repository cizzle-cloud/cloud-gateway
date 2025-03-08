package handlers

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

func ProxyRequestHandler(c *gin.Context, target string) {
	// TODO: move targetURL outside from ProxyRequestHandler and
	// handle error
	targetURL, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	log.Printf("[ PROXY ] Request received at %s at %s\n", c.Request.URL, time.Now())

	c.Request.URL.Host = targetURL.Host
	c.Request.URL.Scheme = targetURL.Scheme
	c.Request.Header.Set("X-Forwarded-Host", c.Request.Header.Get("Host"))
	c.Request.Host = targetURL.Host

	c.Request.URL.Path = targetURL.Path + c.Param("path")

	log.Printf("[ PROXY ] Redirecting request to %s at %s\n", c.Request.URL, time.Now())
	proxy.ServeHTTP(c.Writer, c.Request)
}

func Handle404(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"message": "Page Not Found"})
}
