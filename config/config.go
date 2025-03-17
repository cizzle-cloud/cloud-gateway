package config

import (
	"api_gateway/errors"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type ConfigMiddlewareGroup []string

type ConfigPath struct {
	Method         string `json:"method" yaml:"method"`
	Path           string `json:"path" yaml:"path"`
	ProxyTarget    string `json:"proxy_target,omitempty" yaml:"proxy_target,omitempty"`
	RedirectTarget string `json:"redirect_target,omitempty" yaml:"redirect_target,omitempty"`
}

type ConfigRoute struct {
	Prefix          string       `json:"prefix" yaml:"prefix"`
	Method          string       `json:"method" yaml:"method"`
	Middleware      []string     `json:"middleware,omitempty" yaml:"middleware,omitempty"`
	MiddlewareGroup string       `json:"middleware_group,omitempty" yaml:"middleware_group,omitempty"`
	ProxyTarget     string       `json:"proxy_target,omitempty" yaml:"proxy_target,omitempty"`
	RedirectTarget  string       `json:"redirect_target,omitempty" yaml:"redirect_target,omitempty"`
	Paths           []ConfigPath `json:"paths,omitempty" yaml:"paths,omitempty"`
}

type ConfigDomainRoute struct {
	Domain          string   `json:"domain" yaml:"domain"`
	ProxyTarget     string   `json:"proxy_target" yaml:"proxy_target"`
	Middleware      []string `json:"middleware,omitempty" yaml:"middleware,omitempty"`
	MiddlewareGroup string   `json:"middleware_group,omitempty" yaml:"middleware_group,omitempty"`
}

type Env struct {
	Host                    string `json:"HOST" yaml:"HOST"`
	Port                    string `json:"PORT" yaml:"PORT"`
	ValidateAuthURL         string `json:"VALIDATE_AUTH_URL" yaml:"VALIDATE_AUTH_URL"`
	RedirectUnauthorizedURL string `json:"REDIRECT_UNAUTHORIZED_URL" yaml:"REDIRECT_UNAUTHORIZED_URL"`
}

type Config struct {
	MiddlewareGroups map[string]ConfigMiddlewareGroup `json:"middleware_groups" yaml:"middleware_groups"`
	Routes           []ConfigRoute                    `json:"routes" yaml:"routes"`
	DomainRoutes     []ConfigDomainRoute              `json:"domain_routes" yaml:"domain_routes"`
	Env              Env                              `json:"env" yaml:"env"`
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
	filepath := loadEnvVar("CONFIG_FILEPATH", &errorMsgs) //"config_template.yaml"
	fileType := loadEnvVar("CONFIG_FILETYPE", &errorMsgs) //"yaml"

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
