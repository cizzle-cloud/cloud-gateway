package config

import (
	"api_gateway/errors"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type MiddlewareGroupConfig []string

type RateLimitConfig struct {
	Algorithm string                 `json:"algorithm" yaml:"algorithm"`
	RawConfig map[string]interface{} `json:",inline" yaml:",inline"`
}

type FixedWindowCounterConfig struct {
	Limit      int `json:"limit" yaml:"limit"`
	WindowSize int `json:"window_size" yaml:"window_size"`
}

type TokenBucketConfig struct {
	Capacity       int `json:"capacity" yaml:"capacity"`
	RefillTokens   int `json:"refill_tokens" yaml:"refill_tokens"`
	RefillInterval int `json:"refill_interval" yaml:"refill_interval"`
}

type PathConfig struct {
	Method         string `json:"method" yaml:"method"`
	Path           string `json:"path" yaml:"path"`
	ProxyTarget    string `json:"proxy_target,omitempty" yaml:"proxy_target,omitempty"`
	RedirectTarget string `json:"redirect_target,omitempty" yaml:"redirect_target,omitempty"`
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

type Config struct {
	RateLimiters     map[string]RateLimitConfig       `json:"rate_limiters" yaml:"rate_limiters"`
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

func LoadConfig() (Config, errors.ErrorHandler) {
	var errorMsgs []string
	// filepath := loadEnvVar("CONFIG_FILEPATH", &errorMsgs)
	// fileType := loadEnvVar("CONFIG_FILETYPE", &errorMsgs)
	filepath := "config_template.yaml"
	fileType := "yaml"

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
