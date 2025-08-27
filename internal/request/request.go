package request

import (
	"log"
	"net"
	"net/http"
)

// ClientIP extracts the client IP address from the request's RemoteAddr.
// Falls back to RemoteAddr raw string if parsing fails.
func ClientIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("%v", err)
		return r.RemoteAddr
	}
	return ip
}
