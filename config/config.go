package config

import (
	"api_gateway/errors"
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ConfigMiddlewareGroup []string

// Path defines an individual API route (used in JSON/YAML)
type ConfigPath struct {
	Method         string `json:"method" yaml:"method"`
	Path           string `json:"path" yaml:"path"`
	ProxyTarget    string `json:"proxy_target,omitempty" yaml:"proxy_target,omitempty"`
	RedirectTarget string `json:"redirect_target,omitempty" yaml:"redirect_target,omitempty"`
}

// ConfigRoute represents a route or proxy group from config
type ConfigRoute struct {
	Prefix          string       `json:"prefix" yaml:"prefix"`
	Method          string       `json:"method" yaml:"method"`
	Middleware      []string     `json:"middleware,omitempty" yaml:"middleware,omitempty"`
	MiddlewareGroup string       `json:"middleware_group,omitempty" yaml:"middleware_group,omitempty"`
	ProxyTarget     string       `json:"proxy_target,omitempty" yaml:"proxy_target,omitempty"`
	RedirectTarget  string       `json:"redirect_target,omitempty" yaml:"redirect_target,omitempty"`
	Paths           []ConfigPath `json:"paths,omitempty" yaml:"paths,omitempty"`
}

type Env struct {
	Port                    string `json:"PORT" yaml:"PORT"`
	ValidateAuthURL         string `json:"VALIDATE_AUTH_URL" yaml:"VALIDATE_AUTH_URL"`
	RedirectUnauthorizedURL string `json:"REDIRECT_UNAUTHORIZED_URL" yaml:"REDIRECT_UNAUTHORIZED_URL"`
}

// Config represents the full JSON/YAML structure
type Config struct {
	MiddlewareGroups map[string]ConfigMiddlewareGroup `json:"middleware_groups" yaml:"middleware_groups"`
	Routes           []ConfigRoute                    `json:"routes" yaml:"routes"`
	Env              Env                              `json:"env" yaml:"env"`
}

func LoadConfig(filepath string, fileType string) (Config, errors.ErrorHandler) {
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
