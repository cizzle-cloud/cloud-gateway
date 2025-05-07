package config

import (
	"api_gateway/errors"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type MiddlewareGroupConfig []string

type RateLimitConfig struct {
	Algorithm       string        `json:"algorithm" yaml:"algorithm"`
	Ttl             time.Duration `json:"ttl" yaml:"ttl"`
	CleanupInterval time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`

	Limit      int           `json:"limit" yaml:"limit"`
	WindowSize time.Duration `json:"window_size" yaml:"window_size"`

	Capacity       int           `json:"capacity" yaml:"capacity"`
	RefillTokens   int           `json:"refill_tokens" yaml:"refill_tokens"`
	RefillInterval time.Duration `json:"refill_interval" yaml:"refill_interval"`
}

type PathConfig struct {
	Method          string   `json:"method" yaml:"method"`
	Path            string   `json:"path" yaml:"path"`
	Middleware      []string `json:"middleware" yaml:"middleware"`
	MiddlewareGroup string   `json:"middleware_group" yaml:"middleware_group"`
	ProxyTarget     string   `json:"proxy_target" yaml:"proxy_target"`
	RedirectTarget  string   `json:"redirect_target" yaml:"redirect_target"`
	RedirectCode    int      `json:"redirect_code" yaml:"redirect_code"`
}

type RouteConfig struct {
	Prefix          string        `json:"prefix" yaml:"prefix"`
	Method          string        `json:"method" yaml:"method"`
	Middleware      []string      `json:"middleware" yaml:"middleware"`
	MiddlewareGroup string        `json:"middleware_group" yaml:"middleware_group"`
	ProxyTarget     string        `json:"proxy_target" yaml:"proxy_target"`
	RedirectTarget  string        `json:"redirect_target" yaml:"redirect_target"`
	RedirectCode    int           `json:"redirect_code" yaml:"redirect_code"`
	Paths           []*PathConfig `json:"paths" yaml:"paths"`
}

type DomainRouteConfig struct {
	Domain          string   `json:"domain" yaml:"domain"`
	ProxyTarget     string   `json:"proxy_target" yaml:"proxy_target"`
	Middleware      []string `json:"middleware" yaml:"middleware"`
	MiddlewareGroup string   `json:"middleware_group" yaml:"middleware_group"`
}

type ForwardAuthConfig struct {
	Url                  string        `json:"url" yaml:"url"`
	Method               string        `json:"method" yaml:"method"`
	Timeout              time.Duration `json:"timeout" yaml:"timeout"`
	TrustForwardHeader   bool          `json:"trust_forward_header" yaml:"trust_forward_header"`
	ForwardBody          bool          `json:"forward_body" yaml:"forward_body"`
	RequestHeaders       []string      `json:"request_headers" yaml:"request_headers"`
	ResponseHeaders      []string      `json:"response_headers" yaml:"response_headers"`
	AddCookiesToRequest  []string      `json:"add_cookies_to_request" yaml:"add_cookies_to_request"`
	AddCookiesToResponse []string      `json:"add_cookies_to_response" yaml:"add_cookies_to_response"`
}

type NoCachePolicyConfig struct{}

type EnvConfig struct {
	Host         string `json:"HOST" yaml:"HOST"`
	Port         int    `json:"PORT" yaml:"PORT"`
	CertFilepath string `json:"CERT_FILEPATH" yaml:"CERT_FILEPATH"`
	KeyFilepath  string `json:"KEY_FILEPATH" yaml:"KEY_FILEPATH"`
}

type Config struct {
	RateLimiters     map[string]*RateLimitConfig       `json:"rate_limiters" yaml:"rate_limiters"`
	ForwardAuth      map[string]*ForwardAuthConfig     `json:"forward_auth" yaml:"forward_auth"`
	NoCachePolicies  map[string]*NoCachePolicyConfig   `json:"no_cache_policies" yaml:"no_cache_policies"`
	MiddlewareGroups map[string]*MiddlewareGroupConfig `json:"middleware_groups" yaml:"middleware_groups"`
	Routes           []*RouteConfig                    `json:"routes" yaml:"routes"`
	DomainRoutes     []*DomainRouteConfig              `json:"domain_routes" yaml:"domain_routes"`
	Env              *EnvConfig                        `json:"env" yaml:"env"`
}

func isValidMethod(m string) bool {
	switch m {
	case "GET", "POST", "PUT", "DELETE", "PATCH":
		return true
	default:
		return false
	}
}

func isValidRedirectCode(code int) bool {
	switch code {
	case http.StatusFound,
		http.StatusSeeOther,
		http.StatusTemporaryRedirect,
		http.StatusPermanentRedirect:
		return true
	default:
		return false
	}
}

func (cfg *Config) validate() string {
	for _, routeCfg := range cfg.Routes {
		if errString := routeCfg.validate(); errString != "" {
			return errString
		}
	}

	for _, domainCfg := range cfg.DomainRoutes {
		if errString := domainCfg.validate(); errString != "" {
			return errString
		}
	}

	for _, rateLimiterCfg := range cfg.RateLimiters {
		if errString := rateLimiterCfg.validate(); errString != "" {
			return errString
		}
	}

	for _, forwardAuthCfg := range cfg.ForwardAuth {
		if errString := forwardAuthCfg.validate(); errString != "" {
			return errString
		}
	}

	if errString := cfg.Env.validate(); errString != "" {
		return errString
	}

	return ""
}

type Validatable interface {
	validate() string
}

func (cfg *RouteConfig) validate() string {
	if cfg.Prefix == "" {
		return "prefix is missing for base route"
	}

	if cfg.ProxyTarget != "" {
		if cfg.RedirectTarget != "" {
			return "base route with both 'proxy_target' and 'redirect_target' defined is not allowed"
		}

		if cfg.RedirectCode != 0 {
			return "base route with 'proxy_target' and 'redirect_code' defined is not allowed"
		}
	}

	if cfg.RedirectCode != 0 {
		if cfg.RedirectTarget == "" {
			return "'redirect_code' defined without a corresponding 'redirect_target' in base route"
		}
		if !isValidRedirectCode(cfg.RedirectCode) {
			return fmt.Sprintf("invalid 'redirect_code' %d for base route", cfg.RedirectCode)
		}
	} else {
		if cfg.RedirectTarget != "" {
			return "defining 'redirect_target' in base route without defining 'redirect_code' is not allowed"
		}
	}

	for _, pathCfg := range cfg.Paths {
		if pathCfg.ProxyTarget != "" {
			if pathCfg.RedirectTarget != "" {
				return "path route with both 'proxy_target' and 'redirect_target' defined is not allowed"
			}

			if pathCfg.RedirectCode != 0 {
				return "path route with 'proxy_target' and 'redirect_code' defined is not allowed"
			}
		}

		if pathCfg.Path == "" {
			return fmt.Sprintf("path route under prefix '%s' is missing a 'path'", cfg.Prefix)
		}
	}

	if len(cfg.Paths) != 0 {
		if cfg.ProxyTarget != "" {
			return "base route with defined 'proxy_target' url is not allowed to have paths"
		}

		if cfg.RedirectTarget != "" {
			return "base route with defined 'redirect_target' url is not allowed to have paths"
		}
	}

	if cfg.ProxyTarget == "" && cfg.RedirectTarget == "" {
		if len(cfg.Paths) == 0 {
			return "'proxy_target' or 'redirect_target' url is missing for route with no paths"
		}

		if cfg.RedirectCode != 0 {
			return "defining 'redirect_code' in base route that has paths is not allowed"
		}

		for _, pathCfg := range cfg.Paths {
			if pathCfg.ProxyTarget == "" && pathCfg.RedirectTarget == "" {
				return "found base route with path route that have both no 'proxy_target' or 'redirect_target' defined"
			}

			if pathCfg.RedirectCode != 0 {
				if pathCfg.RedirectTarget == "" {
					return "'redirect_code' defined without a corresponding 'redirect_target' in path route"
				}
				if !isValidRedirectCode(pathCfg.RedirectCode) {
					return fmt.Sprintf("invalid 'redirect_code' %d for path route", pathCfg.RedirectCode)
				}
			} else {
				if pathCfg.RedirectTarget != "" {
					return "defining 'redirect_target' in path route without defining 'redirect_code' is not allowed"
				}
			}

		}
	}

	if len(cfg.Paths) == 0 && cfg.Method == "" {
		return "http method is missing for a route with no paths"
	}

	if cfg.Method != "" {
		if !isValidMethod(cfg.Method) {
			return fmt.Sprintf("found invalid http method '%s' in a route", cfg.Method)
		}
		for _, pathCfg := range cfg.Paths {
			if pathCfg.Method != "" {
				return "http method should not be specified both at route and path level"
			}
		}
	}

	if cfg.Method == "" {
		for _, pathCfg := range cfg.Paths {
			if pathCfg.Method == "" {
				return fmt.Sprintf(
					"path '%s' has no http method and its base route also has no method",
					pathCfg.Path,
				)
			}

			if !isValidMethod(pathCfg.Method) {
				return fmt.Sprintf("found invalid http method '%s' in a route path", pathCfg.Method)
			}
		}
	}

	return ""
}

func (cfg *DomainRouteConfig) validate() string {
	if cfg.Domain == "" {
		return "field 'domain' is missing for domain route"
	}

	if cfg.ProxyTarget == "" {
		return "field 'proxy_target' is missing for domain route"
	}

	return ""
}

func (cfg *RateLimitConfig) validate() string {
	switch algo := cfg.Algorithm; algo {
	case "":
		return "'algorithm' field is not specified for rate limiter"
	case "fixed_window_counter":
		if cfg.Capacity != 0 {
			return "wrong option 'capacity' is specified for rate limiter 'fixed_window_counter'"
		}

		if cfg.RefillTokens != 0 {
			return "wrong option 'refill_tokens' is specified for rate limiter 'fixed_window_counter'"
		}

		if cfg.RefillInterval != 0 {
			return "wrong option 'refill_interval' is specified for rate limiter 'fixed_window_counter'"
		}

		if cfg.Limit <= 0 {
			return "'limit' must be a positive integer"
		}
	case "token_bucket":
		if cfg.Limit != 0 {
			return "wrong option 'limit' is specified for rate limiter 'token_bucket'"
		}

		if cfg.WindowSize != 0 {
			return "wrong option 'window_size' is specified for rate limiter 'token_bucket'"
		}

		if cfg.Capacity <= 0 {
			return "'capacity' must be a positive integer"
		}

		if cfg.RefillTokens <= 0 {
			return "'refill_tokens' must be a positive integer"
		}
	default:
		return fmt.Sprintf("unknown rate limit algorithm '%s' specified", algo)
	}
	return ""
}

func (cfg *ForwardAuthConfig) validate() string {
	if cfg.Url == "" {
		return "required field 'url' is missing for forward auth middleware"
	}

	return ""
}

func (cfg *EnvConfig) validate() string {
	if cfg.Port < 0 || cfg.Port > 65535 {
		return "invalid 'PORT'. Port number must be in the range of 0-65535"
	}
	return ""
}

func (cfg *Config) setDefaults() {
	for _, forwardAuthCfg := range cfg.ForwardAuth {
		forwardAuthCfg.setDefaults()
	}

	for _, routeCfg := range cfg.Routes {
		routeCfg.setDefaults()
	}

	cfg.Env.setDefaults()
}

func (cfg *ForwardAuthConfig) setDefaults() {
	if cfg.Method == "" {
		cfg.Method = "GET"
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
}

func (cfg *RouteConfig) setDefaults() {
	for _, pathCfg := range cfg.Paths {
		if pathCfg.Method == "" {
			pathCfg.Method = cfg.Method
		}
	}
}

func (cfg *EnvConfig) setDefaults() {
	if cfg.Port == 0 {
		cfg.Port = 8080
	}

	if cfg.Host == "" {
		cfg.Host = "0.0.0.0"
	}
}

func loadEnvVar(key string, errorMsgs *[]string) string {
	value := os.Getenv(key)
	if value == "" {
		*errorMsgs = append(*errorMsgs, fmt.Sprintf("%s is not set", key))
	}
	return value
}

type Env struct {
	ConfigFilepath string
	ConfigFileType string
}

func LoadEnv() (Env, errors.ErrorHandler) {
	var errorMsgs []string

	configFilepath := loadEnvVar("CONFIG_FILEPATH", &errorMsgs)
	fileExt := filepath.Ext(configFilepath)
	fileType := strings.TrimPrefix(fileExt, ".")

	if len(errorMsgs) > 0 {
		return Env{}, &errors.LoadConfigError{
			Message: "error while loading Env Vars:\n" + strings.Join(errorMsgs, "\n"),
		}
	}

	return Env{ConfigFilepath: configFilepath, ConfigFileType: fileType}, nil
}

func LoadConfig(filepath, fileType string) (*Config, errors.ErrorHandler) {
	var errorMsgs []string

	if len(errorMsgs) > 0 {
		return &Config{}, &errors.LoadConfigError{
			Message: "error while loading Env Vars:\n" + strings.Join(errorMsgs, "\n"),
		}
	}

	file, err := os.ReadFile(filepath)
	if err != nil {
		return &Config{}, &errors.LoadConfigError{
			Message: fmt.Sprintf("error while reading config file: %v", err),
		}
	}

	var cfg Config
	switch fileType {
	case "yaml":
		err = yaml.Unmarshal(file, &cfg)
	case "json":
		err = json.Unmarshal(file, &cfg)
	default:
		err = fmt.Errorf("unknown file type %s", fileType)
	}

	if err != nil {
		return &Config{}, &errors.LoadConfigError{
			Message: fmt.Sprintf("error while unmarshaling config file: %v", err),
		}
	}

	cfg.setDefaults()

	if errString := cfg.validate(); errString != "" {
		return &Config{}, &errors.LoadConfigError{
			Message: errString,
		}
	}

	return &cfg, nil
}
