package request

import (
	"log"
	"net"
	"net/http"
	"strings"
)

type Context struct {
	trustedProxies []string
	trustHeaders   []string
	trustedCIDRs   []*net.IPNet
}

func NewContext(trustedProxies, trustHeaders []string) (*Context, error) {
	trustedCIDRs, err := getTrustedCIDRs(trustedProxies)
	if err != nil {
		return nil, err
	}
	return &Context{
		trustedProxies: trustedProxies,
		trustHeaders:   trustHeaders,
		trustedCIDRs:   trustedCIDRs,
	}, nil
}

func (c *Context) isTrustedProxy(ip net.IP) bool {
	if c.trustedCIDRs == nil {
		return false
	}
	for _, cidr := range c.trustedCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

// validateHeader will parse a header value and return the client IP address
// and a boolean indicating if header is valid.
func (c *Context) validateHeader(value string) (string, bool) {
	if value == "" {
		return "", false
	}
	ipStr := strings.TrimSpace(value)
	ip := net.ParseIP(ipStr)
	if c.isTrustedProxy(ip) {
		return ip.String(), true
	}
	return "", false
}

// getTrustedCIDRs converts a list of IPs or CIDR strings into a slice of *net.IPNet.
// Plain IPs are converted to single-host CIDRs (/32 for IPv4, /128 for IPv6).
func getTrustedCIDRs(trustedProxies []string) ([]*net.IPNet, error) {
	if trustedProxies == nil {
		return nil, nil
	}

	cidr := make([]*net.IPNet, 0, len(trustedProxies))

	for _, proxy := range trustedProxies {
		if !strings.Contains(proxy, "/") {
			ip := parseIP(proxy)
			if ip == nil {
				return cidr, &net.ParseError{Type: "IP address", Text: proxy}
			}

			switch len(ip) {
			case net.IPv4len:
				proxy += "/32"
			case net.IPv6len:
				proxy += "/128"
			}
		}
		_, cidrNet, err := net.ParseCIDR(proxy)
		if err != nil {
			return cidr, err
		}
		cidr = append(cidr, cidrNet)
	}
	return cidr, nil
}

// parseIP parses a string representation of an IP address and returns a net.IP
// in its canonical byte form. That is 4-byte slice for IPv4 and 16-byte slice for IPv6.
// Returns nil if the input is invalid.
func parseIP(ip string) net.IP {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil
	}

	if ipv4 := parsedIP.To4(); ipv4 != nil {
		return ipv4
	}

	return parsedIP
}

// RemoteIP parses the IP from r.RemoteAddr, normalizes and returns the IP without the port.
func RemoteIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		log.Printf("%v", err)
		return ""
	}
	return ip
}

// ClientIP returns the real client IP address from the request.
// Falls back to RemoteAddr raw string if parsing fails.
// Check if remoteIP is a trusted proxy or not.
// then try to parse the headers defined in Config, defaulting to [X-Forwarded-For, X-Real-IP].
// If the headers are not syntactically valid OR the remote IP does not correspond to a trusted proxy,
// the remote IP is returned.
func ClientIP(c *Context, r *http.Request) string {
	remoteIP := net.ParseIP(RemoteIP(r))
	if remoteIP == nil {
		return ""
	}
	trusted := c.isTrustedProxy(remoteIP)
	if trusted && c.trustHeaders != nil {
		for _, headerName := range c.trustHeaders {
			headerValue := r.Header.Get(headerName)
			ip, valid := c.validateHeader(headerValue)
			if valid {
				return ip
			}
		}
	}
	return remoteIP.String()
}
