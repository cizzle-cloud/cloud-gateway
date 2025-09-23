package request

import (
	"net"
	"net/http"
	"testing"
)

func TestNewContext(t *testing.T) {
	tests := []struct {
		name           string
		trustedProxies []string
		trustedHeaders []string
		expectError    bool
	}{
		{
			name:           "valid proxies and headers",
			trustedProxies: []string{"192.168.1.1", "10.0.0.0/8"},
			trustedHeaders: []string{"X-Forwarded-For", "X-Real-IP"},
			expectError:    false,
		},
		{
			name:           "empty slices",
			trustedProxies: []string{},
			trustedHeaders: []string{},
			expectError:    false,
		},
		{
			name:           "nil slices",
			trustedProxies: nil,
			trustedHeaders: nil,
			expectError:    false,
		},
		{
			name:           "invalid IP address",
			trustedProxies: []string{"invalid-ip"},
			trustedHeaders: []string{"X-Forwarded-For", "X-Real-IP"},
			expectError:    true,
		},
		{
			name:           "invalid CIDR",
			trustedProxies: []string{"10.0.0.0/33"},
			trustedHeaders: []string{"X-Forwarded-For", "X-Real-IP"},
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(t.Name(), func(t *testing.T) {
			c, err := NewContext(tt.trustedProxies, tt.trustedHeaders)
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if c == nil {
					t.Error("expected context but got nil")
				}
			}
		})
	}
}

func TestContext_isTrustedProxy(t *testing.T) {
	c, err := NewContext([]string{"192.168.1.1", "10.0.0.0/8", "::1"}, nil)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	cNoProxies, _ := NewContext(nil, nil)
	if err != nil {
		t.Fatalf("failed to create no proxies context: %v", err)
	}

	tests := []struct {
		name     string
		ip       string
		context  *Context
		expected bool
	}{
		{
			name:     "trusted IPv4 single-host",
			ip:       "192.168.1.1",
			context:  c,
			expected: true,
		},
		{
			name:     "trusted IPv4 in CIDR range",
			ip:       "10.255.255.255",
			context:  c,
			expected: true,
		},
		{
			name:     "trusted IPv6",
			ip:       "::1",
			context:  c,
			expected: true,
		},
		{
			name:     "untrusted IPv4",
			ip:       "3.3.3.3",
			context:  c,
			expected: false,
		},
		{
			name:     "untrusted IPv6",
			ip:       "2001:db8::1",
			context:  c,
			expected: false,
		},
		{
			name:     "no proxies context",
			ip:       "192.168.1.1",
			context:  cNoProxies,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("failed to parse IP: %s", tt.ip)
			}
			result := tt.context.isTrustedProxy(ip)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestContext_validateHeader(t *testing.T) {
	c, err := NewContext([]string{"192.168.1.1", "10.0.0.0/8"}, nil)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	tests := []struct {
		name        string
		value       string
		expectedIP  string
		expectValid bool
	}{
		{
			name:        "valid IP",
			value:       "192.168.1.1",
			expectedIP:  "192.168.1.1",
			expectValid: true,
		},
		{
			name:        "valid IP with whitespace",
			value:       "  10.255.255.255  ",
			expectedIP:  "10.255.255.255",
			expectValid: true,
		},
		{
			name:        "untrusted IP",
			value:       "8.8.8.8",
			expectedIP:  "",
			expectValid: false,
		},
		{
			name:        "invalid IP",
			value:       "invalid-ip",
			expectedIP:  "",
			expectValid: false,
		},
		{
			name:        "empty string",
			value:       "",
			expectedIP:  "",
			expectValid: false,
		},
		{
			name:        "whitespace only",
			value:       "   ",
			expectedIP:  "",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip, valid := c.validateHeader(tt.value)
			if ip != tt.expectedIP {
				t.Errorf("expected IP %s, got %s", tt.expectedIP, ip)
			}
			if valid != tt.expectValid {
				t.Errorf("expected valid %v, got %v", tt.expectValid, valid)
			}
		})
	}
}

func TestGetTrustedCIDRs(t *testing.T) {
	tests := []struct {
		name           string
		trustedProxies []string
		expectedCount  int
		expectError    bool
	}{
		{
			name:           "nil slice",
			trustedProxies: nil,
			expectedCount:  0,
			expectError:    false,
		},
		{
			name:           "empty slice",
			trustedProxies: []string{},
			expectedCount:  0,
			expectError:    false,
		},
		{
			name:           "IPv4 addresses",
			trustedProxies: []string{"192.168.1.1", "10.0.0.1"},
			expectedCount:  2,
			expectError:    false,
		},
		{
			name:           "IPv6 addresses",
			trustedProxies: []string{"::1", "2001:db8::1"},
			expectedCount:  2,
			expectError:    false,
		},
		{
			name:           "CIDR ranges",
			trustedProxies: []string{"192.168.0.0/24", "10.0.0.0/8"},
			expectedCount:  2,
			expectError:    false,
		},
		{
			name:           "mixed IPs and CIDRs",
			trustedProxies: []string{"192.168.1.1", "10.0.0.0/8", "::1"},
			expectedCount:  3,
			expectError:    false,
		},
		{
			name:           "invalid IP",
			trustedProxies: []string{"192.168.1.1", "invalid-ip"},
			expectedCount:  0,
			expectError:    true,
		},
		{
			name:           "invalid CIDR",
			trustedProxies: []string{"192.168.1.1", "192.168.1.1/33"},
			expectedCount:  0,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cidrs, err := getTrustedCIDRs(tt.trustedProxies)
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(cidrs) != tt.expectedCount {
					t.Errorf("expected %d CIDRs, got %d", tt.expectedCount, len(cidrs))
				}
			}
		})
	}
}

func TestParseIP(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		expectedIP string
	}{
		{
			name:       "valid IPv4",
			value:      "192.168.1.1",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "valid IPv6",
			value:      "2001:db8::1",
			expectedIP: "2001:db8::1",
		},
		{
			name:       "IPv4-mapped IPv6 should return IPv4",
			value:      "::ffff:192.168.1.1",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "invalid IP",
			value:      "invalid-ip",
			expectedIP: "<nil>",
		},
		{
			name:       "empty string",
			value:      "",
			expectedIP: "<nil>",
		},
		{
			name:       "out of range IPv4",
			value:      "256.256.256.256",
			expectedIP: "<nil>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseIP(tt.value).String()
			if result != tt.expectedIP {
				t.Errorf("expected IP %v, got %v", result, tt.expectedIP)
			}
		})
	}
}

func TestRemoteIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		expected   string
	}{
		{
			name:       "IPv4 with port",
			remoteAddr: "192.168.1.1:443",
			expected:   "192.168.1.1",
		},
		{
			name:       "IPv6 with port",
			remoteAddr: "[2001:db8::1]:443",
			expected:   "2001:db8::1",
		},
		{
			name:       "IPv4 with whitespace",
			remoteAddr: "  192.168.1.1:443  ",
			expected:   "192.168.1.1",
		},
		{
			name:       "invalid address",
			remoteAddr: "invalid-address",
			expected:   "",
		},
		{
			name:       "empty string",
			remoteAddr: "",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				RemoteAddr: tt.remoteAddr,
			}
			result := RemoteIP(req)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestClientIP(t *testing.T) {
	c, err := NewContext(
		[]string{"192.168.0.0/24", "10.0.0.0/8"},
		[]string{"X-Forwarded-For", "X-Real-IP"},
	)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	cNoHeaders, err := NewContext([]string{"192.168.0.0/24", "10.0.0.0/8"}, nil)
	if err != nil {
		t.Fatalf("failed to create no headers context: %v", err)
	}

	tests := []struct {
		name        string
		context     *Context
		remoteAddr  string
		headerName  string
		headerValue string
		expected    string
	}{
		{
			name:        "trusted proxy with X-Forwarded-For",
			context:     c,
			remoteAddr:  "192.168.0.100:443",
			headerName:  "X-Forwarded-For",
			headerValue: "192.168.0.50",
			expected:    "192.168.0.50",
		},
		{
			name:        "trusted proxy with X-Real-IP",
			context:     c,
			remoteAddr:  "10.0.0.5:443",
			headerName:  "X-Real-IP",
			headerValue: "10.0.10.50",
			expected:    "10.0.10.50",
		},
		{
			name:        "trusted proxy with invalid header name",
			context:     c,
			remoteAddr:  "10.0.0.5:443",
			headerName:  "Invalid-Header-Name",
			headerValue: "10.0.10.50",
			expected:    "10.0.0.5",
		},
		{
			name:        "trusted proxy with invalid header value",
			context:     c,
			remoteAddr:  "10.0.0.5:443",
			headerName:  "X-Real-IP",
			headerValue: "3.0.0.5",
			expected:    "10.0.0.5",
		},
		{
			name:        "trusted proxy with empty header value",
			context:     c,
			remoteAddr:  "10.0.0.5:443",
			headerName:  "X-Real-IP",
			headerValue: "",
			expected:    "10.0.0.5",
		},
		{
			name:        "trusted proxy with whitespace header value",
			context:     c,
			remoteAddr:  "10.0.0.5:443",
			headerName:  "X-Real-IP",
			headerValue: "	",
			expected:    "10.0.0.5",
		},
		{
			name:        "untrusted proxy",
			context:     c,
			remoteAddr:  "3.0.0.5:443",
			headerName:  "X-Real-IP",
			headerValue: "10.0.10.50",
			expected:    "3.0.0.5", // ignore headers for unstrusted proxy
		},
		{
			name:        "invalid remote address",
			context:     c,
			remoteAddr:  "invalid-address",
			headerName:  "X-Real-IP",
			headerValue: "10.0.10.50",
			expected:    "",
		},
		{
			name:        "trusted proxy with no trust headers context",
			context:     cNoHeaders,
			remoteAddr:  "192.168.0.100:443",
			headerName:  "X-Forwarded-For",
			headerValue: "192.168.0.50",
			expected:    "192.168.0.100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				RemoteAddr: tt.remoteAddr,
				Header:     make(http.Header),
			}

			req.Header.Set(tt.headerName, tt.headerValue)

			result := ClientIP(tt.context, req)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
