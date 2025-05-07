package middleware

import (
	"api_gateway/config"
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestForwardAuthMiddlewareAuthorized(t *testing.T) {
	authorizationHeader := "Bearer test123"
	forwardBody := `{"k1": "v1", "k2": "v2"}`
	method := "GET"

	trustForwardHeaderTest := []struct {
		label    string
		header   string
		expected string
	}{
		{label: "Authorization", header: "Authorization", expected: authorizationHeader},
		{label: "Mock-Header", header: "Mock-Header", expected: "mock-header"},
		{label: "Mock-Header-2", header: "Mock-Header-2", expected: ""},
		{label: "host", header: "X-Forwarded-Host", expected: "example.com"},
		{label: "method", header: "X-Forwarded-Method", expected: "GET"},
		{label: "uri", header: "X-Forwarded-Uri", expected: "/protected"},
		{label: "for", header: "X-Forwarded-For", expected: "192.0.2.1"},
		{label: "proto", header: "X-Forwarded-Proto", expected: "http"},
	}

	// Mock auth server
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test-Header", "test_header")
		w.Header().Set("X-Test-Header-2", "test_header_2")
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "abc123"})
		http.SetCookie(w, &http.Cookie{Name: "csrf", Value: "efg456"})
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"authorized"}`))

		for _, tt := range trustForwardHeaderTest {
			actual := r.Header.Get(tt.header)
			if tt.expected != actual {
				t.Errorf("Expected %s: %s, got %s", tt.label, tt.expected, actual)

			}
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		actual := string(bodyBytes)
		if forwardBody != actual {
			t.Errorf("Expected body: %s, got %s", forwardBody, actual)
		}

		if r.Method != method {
			t.Errorf("Expected http method: %s, got %s", method, r.Method)
		}
	}))
	defer authServer.Close()

	cfg := config.ForwardAuthConfig{
		Url:                  authServer.URL,
		Timeout:              2 * time.Second,
		Method:               method,
		ForwardBody:          true,
		TrustForwardHeader:   true,
		RequestHeaders:       []string{"Authorization", "Mock-Header"},
		ResponseHeaders:      []string{"X-Test-Header"},
		AddCookiesToResponse: []string{"session"},
	}

	// Gin test setup
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(NewForwardAuthMiddleware(&cfg))
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	// Mock request
	body := bytes.NewBuffer([]byte(forwardBody))
	req := httptest.NewRequest("GET", "/protected", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authorizationHeader)
	req.Header.Set("Mock-Header", "mock-header")
	req.Header.Set("Mock-Header-2", "mock-header-2")
	w := httptest.NewRecorder()

	// Do request
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code: %v, got %v", http.StatusOK, w.Code)
	}

	actualBody := w.Body.String()
	if w.Body.String() != `{"message":"OK"}` {
		t.Errorf("Expected body string: %s, got %s", `{"message":"OK"}`, actualBody)
	}

	actualCookies := w.Header().Values("Set-Cookie")
	if len(actualCookies) != len(cfg.AddCookiesToResponse) {
		t.Error("Not all cookies should have been added to the response.")
	}

	if actualCookies[0] != "session=abc123" {
		t.Errorf("Expected cookie: %s, got %s", "session=abc123", actualCookies[0])
	}

	testHeader := w.Header().Get("X-Test-Header")
	if testHeader != "test_header" {
		t.Errorf("Expected X-Test-Header: %s, got %s", "test_header", testHeader)
	}
	if w.Header().Get("X-Test-Header-2") != "" {
		t.Error("X-Test-Header-2 should not have been forwarded")
	}
}

func TestForwardAuthMiddlewareUnauthorized(t *testing.T) {
	authorizationHeader := "Bearer test123"
	forwardBody := `{"k1": "v1", "k2": "v2"}`
	method := "GET"

	trustForwardHeaderTest := []struct {
		label    string
		header   string
		expected string
	}{
		{label: "Authorization", header: "Authorization", expected: ""},
		{label: "Mock-Header", header: "Mock-Header", expected: ""},
		{label: "Mock-Header-2", header: "Mock-Header-2", expected: ""},
		{label: "host", header: "X-Forwarded-Host", expected: ""},
		{label: "method", header: "X-Forwarded-Method", expected: ""},
		{label: "uri", header: "X-Forwarded-Uri", expected: ""},
		{label: "for", header: "X-Forwarded-For", expected: ""},
		{label: "proto", header: "X-Forwarded-Proto", expected: ""},
	}

	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test-Header", "test_header")
		w.Header().Set("X-Test-Header-2", "test_header_2")
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "abc123"})
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))

		for _, tt := range trustForwardHeaderTest {
			actual := r.Header.Get(tt.header)
			if tt.expected != actual {
				t.Errorf("Expected %s: %s, got %s", tt.label, tt.expected, actual)

			}
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		actual := string(bodyBytes)
		if actual != "" {
			t.Errorf("No body expected to be forwarded")
		}

		if r.Method != method {
			t.Errorf("Expected http method: %s, got %s", method, r.Method)
		}
	}))
	defer authServer.Close()

	cfg := config.ForwardAuthConfig{
		Url:                authServer.URL,
		Timeout:            2 * time.Second,
		Method:             method,
		ForwardBody:        false,
		TrustForwardHeader: false,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(NewForwardAuthMiddleware(&cfg))
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	body := bytes.NewBuffer([]byte(forwardBody))
	req := httptest.NewRequest("GET", "/protected", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authorizationHeader)
	req.Header.Set("Mock-Header", "mock-header")
	req.Header.Set("Mock-Header-2", "mock-header-2")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code: %v, got %v", http.StatusUnauthorized, w.Code)
	}

	actualBody := w.Body.String()
	if w.Body.String() != `{"error":"unauthorized"}` {
		t.Errorf("Expected body string: %s, got %s", `{"error":"unauthorized"}`, actualBody)
	}

	actualCookies := w.Header().Values("Set-Cookie")

	if len(actualCookies) != 0 {
		t.Error("No cookies should have been added to the response")
	}
	if w.Header().Get("X-Test-Header") != "" {
		t.Error("X-Test-Header should not have been forwarded to the response")
	}
	if w.Header().Get("X-Test-Header-2") != "" {
		t.Error("X-Test-Header-2 should not have been forwarded to the response")
	}
}

func TestForwardAuthMiddlewareTimeout(t *testing.T) {
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"authorized"}`))
	}))
	defer authServer.Close()

	cfg := config.ForwardAuthConfig{
		Url:                authServer.URL,
		Timeout:            1 * time.Second,
		Method:             "GET",
		ForwardBody:        false,
		TrustForwardHeader: false,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(NewForwardAuthMiddleware(&cfg))
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status code: %v, got %v", http.StatusServiceUnavailable, w.Code)
	}

	if !strings.Contains(w.Body.String(), "auth service unreachable") {
		t.Errorf("Expected auth service to be unreachable")
	}
}
