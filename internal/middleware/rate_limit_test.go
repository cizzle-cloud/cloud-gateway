package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cizzle-cloud/cloud-gateway/internal/config"
	"github.com/cizzle-cloud/cloud-gateway/internal/request"
)

func setupRateLimitHandler(t *testing.T, c *request.Context, cfg *config.RateLimitConfig) http.Handler {
	t.Helper()

	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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

func TestRateLimit_Allowed(t *testing.T) {
	c, _ := request.NewContext([]string{}, []string{})
	cfg := config.RateLimitConfig{
		Algorithm:       "token_bucket",
		TTL:             time.Minute,
		CleanupInterval: time.Minute,
		Capacity:        10,
		RefillTokens:    1,
		RefillInterval:  time.Minute,
	}

	handler := setupRateLimitHandler(t, c, &cfg)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/protected", nil)
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

func TestRateLimit_Blocked(t *testing.T) {
	c, _ := request.NewContext([]string{}, []string{})
	cfg := config.RateLimitConfig{
		Algorithm:       "token_bucket",
		TTL:             time.Minute,
		CleanupInterval: time.Minute,
		Capacity:        0,
		RefillTokens:    0,
		RefillInterval:  time.Minute,
	}

	handler := setupRateLimitHandler(t, c, &cfg)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/protected", nil)
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

func TestRateLimit_MultipleClients(t *testing.T) {
	c, _ := request.NewContext([]string{"192.0.2.1", "192.168.1.1", "192.168.1.2"}, []string{"X-Forwarded-For"})
	cfg := config.RateLimitConfig{
		Algorithm:       "token_bucket",
		TTL:             time.Minute,
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
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/protected", nil)
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

func TestRateLimit_Concurrent(t *testing.T) {
	c, _ := request.NewContext([]string{}, []string{})
	cfg := config.RateLimitConfig{
		Algorithm:       "token_bucket",
		TTL:             time.Minute,
		CleanupInterval: time.Minute,
		Capacity:        50,
		RefillTokens:    0,
		RefillInterval:  time.Minute,
	}

	handler := setupRateLimitHandler(t, c, &cfg)

	var allowed atomic.Int32
	var rejected atomic.Int32

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/protected", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			switch w.Code {
			case http.StatusOK:
				allowed.Add(1)
			case http.StatusTooManyRequests:
				rejected.Add(1)
			default:
				t.Errorf("Unexpected status code: %d", w.Code)
			}
		}()
	}
	wg.Wait()

	if allowed.Load()+rejected.Load() != 100 {
		t.Errorf("Expected 100 total responses, got %d", allowed.Load()+rejected.Load())
	}
	if allowed.Load() != 50 {
		t.Errorf("Expected exactly 50 allowed, got %d", allowed.Load())
	}
	if rejected.Load() != 50 {
		t.Errorf("Expected exactly 50 rejected, got %d", rejected.Load())
	}
}
