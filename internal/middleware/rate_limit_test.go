package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cizzle-cloud/cloud-gateway/internal/config"
	"github.com/cizzle-cloud/cloud-gateway/internal/request"
)

func setupRateLimitHandler(t *testing.T, c *request.Context, cfg *config.RateLimitConfig) http.Handler {
	t.Helper()

	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"message":"OK"}`))
		if err != nil {
			t.Errorf("could not write response body: %v", err)
			return
		}
	})

	return NewRateLimitMiddleware(c, cfg)(protectedHandler)
}

func TestRateLimitAllowed(t *testing.T) {
	c, _ := request.NewContext([]string{}, []string{})
	cfg := config.RateLimitConfig{
		Algorithm:       "token_bucket",
		Ttl:             time.Minute,
		CleanupInterval: time.Minute,
		Capacity:        10,
		RefillTokens:    1,
		RefillInterval:  time.Minute,
	}

	handler := setupRateLimitHandler(t, c, &cfg)

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code: %v, got %v", http.StatusOK, w.Code)
	}

	actualBody := w.Body.String()
	expectedBody := `{"message":"OK"}`
	if actualBody != expectedBody {
		t.Errorf("Expected body string: %s, got %s", expectedBody, actualBody)
	}
}

func TestRateLimitBlocked(t *testing.T) {
	c, _ := request.NewContext([]string{}, []string{})
	cfg := config.RateLimitConfig{
		Algorithm:       "token_bucket",
		Ttl:             time.Minute,
		CleanupInterval: time.Minute,
		Capacity:        0,
		RefillTokens:    0,
		RefillInterval:  time.Minute,
	}

	handler := setupRateLimitHandler(t, c, &cfg)

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status code: %v, got %v", http.StatusTooManyRequests, w.Code)
	}

	actualBody := w.Body.String()
	expectedBody := `{"error":"rate limit exceeded"}`
	if actualBody != expectedBody {
		t.Errorf("Expected body string: %s, got %s", expectedBody, actualBody)
	}
}

func TestRateLimitMultipleClients(t *testing.T) {
	c, _ := request.NewContext([]string{"192.0.2.1", "192.168.1.1", "192.168.1.2"}, []string{"X-Forwarded-For"})
	cfg := config.RateLimitConfig{
		Algorithm:       "token_bucket",
		Ttl:             time.Minute,
		CleanupInterval: time.Minute,
		Capacity:        2,
		RefillTokens:    0,
		RefillInterval:  time.Minute,
	}

	handler := setupRateLimitHandler(t, c, &cfg)

	tests := []struct {
		name         string
		clientIP     string
		expectedCode int
	}{
		{
			name:         "first client's first request allowed",
			clientIP:     "192.168.1.1",
			expectedCode: http.StatusOK,
		},
		{
			name:         "first client's second request allowed",
			clientIP:     "192.168.1.1",
			expectedCode: http.StatusOK,
		},
		{
			name:         "first client's third request blocked",
			clientIP:     "192.168.1.1",
			expectedCode: http.StatusTooManyRequests,
		},
		{
			name:         "second client's first request allowed",
			clientIP:     "192.168.1.2",
			expectedCode: http.StatusOK,
		},
		{
			name:         "second client's second request allowed",
			clientIP:     "192.168.1.2",
			expectedCode: http.StatusOK,
		},
		{
			name:         "second client's third request blocked",
			clientIP:     "192.168.1.2",
			expectedCode: http.StatusTooManyRequests,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/protected", nil)
			req.Header.Set("X-Forwarded-For", tt.clientIP)
			t.Logf("%s: remote ip: %s", tt.name, request.RemoteIP(req))
			t.Logf("%s: client ip: %s", tt.name, request.ClientIP(c, req))
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("%s: Expected status code: %v, got %v", tt.name, tt.expectedCode, w.Code)
			}

		})
	}
}
