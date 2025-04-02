package config

import (
	"api_gateway/errors"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type MiddlewareGroupConfig []string

type RateLimitConfig struct {
	Algorithm       string        `json:"algorithm" yaml:"algorithm"`
	Ttl             time.Duration `json:"ttl" yaml:"ttl"`
	CleanupInterval time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`

	Limit      int           `json:"limit,omitempty" yaml:"limit,omitempty"`
	WindowSize time.Duration `json:"window_size,omitempty" yaml:"window_size,omitempty"`

	Capacity       int           `json:"capacity,omitempty" yaml:"capacity,omitempty"`
	RefillTokens   int           `json:"refill_tokens,omitempty" yaml:"refill_tokens,omitempty"`
	RefillInterval time.Duration `json:"refill_interval,omitempty" yaml:"refill_interval,omitempty"`
}

type PathConfig struct {
	Method          string   `json:"method" yaml:"method"`
	Path            string   `json:"path" yaml:"path"`
	Middleware      []string `json:"middleware,omitempty" yaml:"middleware,omitempty"`
	MiddlewareGroup string   `json:"middleware_group,omitempty" yaml:"middleware_group,omitempty"`
	ProxyTarget     string   `json:"proxy_target,omitempty" yaml:"proxy_target,omitempty"`
	RedirectTarget  string   `json:"redirect_target,omitempty" yaml:"redirect_target,omitempty"`
}

type RouteConfig struct {
	Prefix          string       `json:"prefix" yaml:"prefix"`
	Method          string       `json:"method" yaml:"method"`
	Middleware      []string     `json:"middleware,omitempty" yaml:"middleware,omitempty"`
	MiddlewareGroup string       `json:"middleware_group,omitempty" yaml:"middleware_group,omitempty"`
	ProxyTarget     string       `json:"proxy_target,omitempty" yaml:"proxy_target,omitempty"`
	RedirectTarget  string       `json:"redirect_target,omitempty" yaml:"redirect_target,omitempty"`
	Paths           []PathConfig `json:"paths,omitempty" yaml:"paths,omitempty"`
}

type DomainRouteConfig struct {
	Domain          string   `json:"domain" yaml:"domain"`
	ProxyTarget     string   `json:"proxy_target" yaml:"proxy_target"`
	Middleware      []string `json:"middleware,omitempty" yaml:"middleware,omitempty"`
	MiddlewareGroup string   `json:"middleware_group,omitempty" yaml:"middleware_group,omitempty"`
}

type EnvConfig struct {
	Host                    string `json:"HOST" yaml:"HOST"`
	Port                    string `json:"PORT" yaml:"PORT"`
	ValidateAuthURL         string `json:"VALIDATE_AUTH_URL" yaml:"VALIDATE_AUTH_URL"`
	RedirectUnauthorizedURL string `json:"REDIRECT_UNAUTHORIZED_URL" yaml:"REDIRECT_UNAUTHORIZED_URL"`
}

type AuthConfig struct{}

type NoCachePolicyConfig struct{}

type Config struct {
	RateLimiters     map[string]RateLimitConfig       `json:"rate_limiters" yaml:"rate_limiters"`
	Auth             map[string]AuthConfig            `json:"auth" yaml:"auth"`
	NoCachePolicies  map[string]NoCachePolicyConfig   `json:"no_cache_policies" yaml:"no_cache_policies"`
	MiddlewareGroups map[string]MiddlewareGroupConfig `json:"middleware_groups" yaml:"middleware_groups"`
	Routes           []RouteConfig                    `json:"routes" yaml:"routes"`
	DomainRoutes     []DomainRouteConfig              `json:"domain_routes" yaml:"domain_routes"`
	Env              EnvConfig                        `json:"env" yaml:"env"`
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
	filepath := loadEnvVar("CONFIG_FILEPATH", &errorMsgs)
	fileType := loadEnvVar("CONFIG_FILETYPE", &errorMsgs)
	if len(errorMsgs) > 0 {
		return Env{}, &errors.LoadConfigError{
			Message: "error while loading Env Vars:\n" + strings.Join(errorMsgs, "\n"),
		}
	}

	return Env{ConfigFilepath: filepath, ConfigFileType: fileType}, nil
}

func LoadConfig(filepath, fileType string) (Config, errors.ErrorHandler) {
	var errorMsgs []string

	if len(errorMsgs) > 0 {
		return Config{}, &errors.LoadConfigError{
			Message: "error while loading Env Vars:\n" + strings.Join(errorMsgs, "\n"),
		}
	}

	file, err := os.ReadFile(filepath)
	if err != nil {
		return Config{}, &errors.LoadConfigError{
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
		return Config{}, &errors.LoadConfigError{
			Message: fmt.Sprintf("error while unmarshaling config file: %v", err),
		}
	}

	return cfg, nil
}
