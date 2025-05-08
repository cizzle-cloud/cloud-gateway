package config

import (
	"testing"
	"time"
)

func TestValidate(t *testing.T) {
	testCases := []struct {
		name        string
		cfg         Validatable
		expectedErr string
	}{
		{
			name:        "missing prefix",
			cfg:         &RouteConfig{},
			expectedErr: "prefix is missing for base route",
		},
		{
			name: "both proxy and redirect at base",
			cfg: &RouteConfig{
				Prefix:         "/foo",
				ProxyTarget:    "https://proxy.com",
				RedirectTarget: "https://redirect.com",
			},
			expectedErr: "base route with both 'proxy_target' and 'redirect_target' defined is not allowed",
		},
		{
			name: "both proxy and redirect at route path",
			cfg: &RouteConfig{
				Prefix: "/foo",
				Paths: []*PathConfig{
					{
						Path:           "/bar",
						Method:         "GET",
						ProxyTarget:    "https://proxy.com",
						RedirectTarget: "https://redirect.com",
					},
				},
			},
			expectedErr: "path route with both 'proxy_target' and 'redirect_target' defined is not allowed",
		},
		{
			name: "path route missing 'path' field",
			cfg: &RouteConfig{
				Prefix: "/foo",
				Paths: []*PathConfig{
					{Path: "", Method: "GET"},
				},
			},
			expectedErr: "path route under prefix '/foo' is missing a 'path'",
		},
		{
			name: "base route with defined 'proxy_target' has paths",
			cfg: &RouteConfig{
				Prefix:      "/foo",
				ProxyTarget: "https://proxy.com",
				Paths: []*PathConfig{
					{Path: "/quz", Method: "GET"},
				},
			},
			expectedErr: "base route with defined 'proxy_target' url is not allowed to have paths",
		},
		{
			name: "base route with defined 'redirect_target' has paths",
			cfg: &RouteConfig{
				Prefix:         "/foo",
				RedirectTarget: "https://redirect.com",
				RedirectCode:   307,
				Paths: []*PathConfig{
					{Path: "/quz", Method: "GET"},
				},
			},
			expectedErr: "base route with defined 'redirect_target' url is not allowed to have paths",
		},
		{
			name: "base route has no defined target and paths",
			cfg: &RouteConfig{
				Prefix: "/foo",
				Paths:  []*PathConfig{},
			},
			expectedErr: "'proxy_target' or 'redirect_target' url is missing for route with no paths",
		},
		{
			name: "base route with no target that has a path with no target",
			cfg: &RouteConfig{
				Prefix: "/foo",
				Paths: []*PathConfig{
					{Path: "/quz", Method: "GET"},
				},
			},
			expectedErr: "found base route with path route that have both no 'proxy_target' or 'redirect_target' defined",
		},
		{
			name: "missing method in base and paths",
			cfg: &RouteConfig{
				Prefix:      "/foo",
				ProxyTarget: "https://bar.com",
			},
			expectedErr: "http method is missing for a route with no paths",
		},
		{
			name: "method in both base and path",
			cfg: &RouteConfig{
				Prefix: "/foo",
				Method: "GET",
				Paths: []*PathConfig{
					{Path: "/bar", Method: "POST", ProxyTarget: "https://x.com"},
				},
			},
			expectedErr: "http method should not be specified both at route and path level",
		},
		{
			name: "invalid method in route",
			cfg: &RouteConfig{
				Prefix:      "/foo",
				ProxyTarget: "https://proxy.com",
				Method:      "INVALID",
			},
			expectedErr: "found invalid http method 'INVALID' in a route",
		},
		{
			name: "invalid method in path",
			cfg: &RouteConfig{
				Prefix: "/foo",
				Paths: []*PathConfig{
					{Path: "/bar", Method: "INVALID", ProxyTarget: "https://x.com"},
				},
			},
			expectedErr: "found invalid http method 'INVALID' in a route path",
		},
		{
			name: "path and its base route have no http method",
			cfg: &RouteConfig{
				Prefix: "/foo",
				Paths: []*PathConfig{
					{Path: "/bar", ProxyTarget: "https://x.com"},
				},
			},
			expectedErr: "path '/bar' has no http method and its base route also has no method",
		},
		{
			name: "domain route missing domain field",
			cfg: &DomainRouteConfig{
				ProxyTarget: "https://proxy.com",
			},
			expectedErr: "field 'domain' is missing for domain route",
		},
		{
			name: "domain route missing proxy target field",
			cfg: &DomainRouteConfig{
				Domain: "www.example.com",
			},
			expectedErr: "field 'proxy_target' is missing for domain route",
		},
		{
			name:        "rate limiter missing 'algorithm' field",
			cfg:         &RateLimitConfig{},
			expectedErr: "'algorithm' field is not specified for rate limiter",
		},
		{
			name:        "rate limiter has an invalid 'algorithm' type",
			cfg:         &RateLimitConfig{Algorithm: "INVALID"},
			expectedErr: "unknown rate limit algorithm 'INVALID' specified",
		},
		{
			name: "fixed window counter algorithm has invalid 'capacity' field",
			cfg: &RateLimitConfig{
				Algorithm:  "fixed_window_counter",
				Limit:      10,
				WindowSize: 5 * time.Second,
				Capacity:   2,
			},
			expectedErr: "wrong option 'capacity' is specified for rate limiter 'fixed_window_counter'",
		},
		{
			name: "fixed window counter algorithm has invalid 'refill_tokens' field",
			cfg: &RateLimitConfig{
				Algorithm:    "fixed_window_counter",
				Limit:        10,
				WindowSize:   5 * time.Second,
				RefillTokens: 2,
			},
			expectedErr: "wrong option 'refill_tokens' is specified for rate limiter 'fixed_window_counter'",
		},
		{
			name: "fixed window counter algorithm has invalid 'refill_interval' field",
			cfg: &RateLimitConfig{
				Algorithm:      "fixed_window_counter",
				Limit:          10,
				WindowSize:     5 * time.Second,
				RefillInterval: 5 * time.Second,
			},
			expectedErr: "wrong option 'refill_interval' is specified for rate limiter 'fixed_window_counter'",
		},
		{
			name: "fixed window counter algorithm has negative limit",
			cfg: &RateLimitConfig{
				Algorithm:  "fixed_window_counter",
				Limit:      -10,
				WindowSize: 5 * time.Second,
			},
			expectedErr: "'limit' must be a positive integer",
		},
		{
			name: "token bucket algorithm has invalid 'limit' field",
			cfg: &RateLimitConfig{
				Algorithm:      "token_bucket",
				Capacity:       10,
				RefillTokens:   10,
				RefillInterval: 5 * time.Second,
				Limit:          10,
			},
			expectedErr: "wrong option 'limit' is specified for rate limiter 'token_bucket'",
		},
		{
			name: "token bucket algorithm has invalid 'window_size' field",
			cfg: &RateLimitConfig{
				Algorithm:      "token_bucket",
				Capacity:       10,
				RefillTokens:   10,
				RefillInterval: 5 * time.Second,
				WindowSize:     5 * time.Second,
			},
			expectedErr: "wrong option 'window_size' is specified for rate limiter 'token_bucket'",
		},
		{
			name: "token bucket algorithm has negative capacity",
			cfg: &RateLimitConfig{
				Algorithm:      "token_bucket",
				Capacity:       -10,
				RefillTokens:   10,
				RefillInterval: 5 * time.Second,
			},
			expectedErr: "'capacity' must be a positive integer",
		},
		{
			name: "token bucket algorithm has negative refill tokens",
			cfg: &RateLimitConfig{
				Algorithm:      "token_bucket",
				Capacity:       10,
				RefillTokens:   -10,
				RefillInterval: 5 * time.Second,
			},
			expectedErr: "'refill_tokens' must be a positive integer",
		},
		{
			name:        "forward auth missing 'url' field",
			cfg:         &ForwardAuthConfig{},
			expectedErr: "required field 'url' is missing for forward auth middleware",
		},
		{
			name: "redirect_code without redirect_target at base",
			cfg: &RouteConfig{
				Prefix:       "/foo",
				Method:       "GET",
				RedirectCode: 307,
			},
			expectedErr: "'redirect_code' defined without a corresponding 'redirect_target' in base route",
		},
		{
			name: "invalid redirect_code at base",
			cfg: &RouteConfig{
				Prefix:         "/foo",
				Method:         "GET",
				RedirectTarget: "https://redirect.com",
				RedirectCode:   400,
			},
			expectedErr: "invalid 'redirect_code' 400 for base route",
		},
		{
			name: "proxy_target with redirect_code at base",
			cfg: &RouteConfig{
				Prefix:       "/foo",
				Method:       "GET",
				ProxyTarget:  "https://proxy.com",
				RedirectCode: 307,
			},
			expectedErr: "base route with 'proxy_target' and 'redirect_code' defined is not allowed",
		},
		{
			name: "'redirect_code' missing when 'redirect_target' is defined in base route",
			cfg: &RouteConfig{
				Prefix:         "/foo",
				Method:         "GET",
				RedirectTarget: "https://redirect.com",
			},
			expectedErr: "defining 'redirect_target' in base route without defining 'redirect_code' is not allowed",
		},
		{
			name: "'redirect_code' missing when 'redirect_target' is defined in path route",
			cfg: &RouteConfig{
				Prefix: "/foo",
				Paths: []*PathConfig{
					{
						Path:           "/bar",
						Method:         "GET",
						RedirectTarget: "https://example.com",
					},
				},
			},
			expectedErr: "defining 'redirect_target' in path route without defining 'redirect_code' is not allowed",
		},
		{
			name: "redirect_code without redirect_target in path",
			cfg: &RouteConfig{
				Prefix: "/foo",
				Paths: []*PathConfig{
					{
						Path:         "/bar",
						Method:       "GET",
						RedirectCode: 307,
					},
				},
			},
			expectedErr: "found base route with path route that have both no 'proxy_target' or 'redirect_target' defined",
		},
		{
			name: "invalid redirect_code in path",
			cfg: &RouteConfig{
				Prefix: "/foo",
				Paths: []*PathConfig{
					{
						Path:           "/bar",
						Method:         "GET",
						RedirectTarget: "https://example.com",
						RedirectCode:   123,
					},
				},
			},
			expectedErr: "invalid 'redirect_code' 123 for path route",
		},
		{
			name: "proxy_target with redirect_code in path",
			cfg: &RouteConfig{
				Prefix: "/foo",
				Paths: []*PathConfig{
					{
						Path:         "/bar",
						Method:       "GET",
						ProxyTarget:  "https://proxy.com",
						RedirectCode: 308,
					},
				},
			},
			expectedErr: "path route with 'proxy_target' and 'redirect_code' defined is not allowed",
		},
		{
			name: "valid proxy route",
			cfg: &RouteConfig{
				Prefix:      "/foo",
				Method:      "GET",
				ProxyTarget: "https://bar.com",
			},
			expectedErr: "",
		},
		{
			name: "valid redirect with paths",
			cfg: &RouteConfig{
				Prefix: "/baz",
				Paths: []*PathConfig{
					{Path: "/baz", Method: "GET", RedirectTarget: "https://qux.com", RedirectCode: 302},
				},
			},
			expectedErr: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.validate()
			if err != tc.expectedErr {
				t.Errorf("got error = %q, expected %q", err, tc.expectedErr)
			}
		})
	}
}
