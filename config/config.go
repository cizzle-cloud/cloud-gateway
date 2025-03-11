package config

import (
	"api_gateway/errors"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AuthURL       string
	AuthPageURL   string
	ApiGatewayURL string
	AuthPagePath  string
	AuthPath      string
}

func loadEnvVar(key string, errorMsgs *[]string) string {
	value := os.Getenv(key)
	if value == "" {
		*errorMsgs = append(*errorMsgs, fmt.Sprintf("%s is not set", key))
	}
	return value
}

func LoadConfig() (*Config, errors.ErrorHandler) {
	var errorMsgs []string

	authAddress := loadEnvVar("AUTH_ADDRESS", &errorMsgs)
	authPageAddress := loadEnvVar("AUTH_PAGE_ADDRESS", &errorMsgs)
	apiGatewayAddress := loadEnvVar("API_GATEWAY_ADDRESS", &errorMsgs)
	authPagePath := loadEnvVar("AUTH_PAGE_PATH", &errorMsgs)
	authPath := loadEnvVar("AUTH_PATH", &errorMsgs)

	if len(errorMsgs) > 0 {
		return nil, &errors.LoadConfigError{
			Message: "error while loading Config:\n" + strings.Join(errorMsgs, "\n"),
		}
	}

	return &Config{
		AuthURL:       authAddress,
		AuthPageURL:   authPageAddress,
		ApiGatewayURL: apiGatewayAddress,
		AuthPagePath:  authPagePath,
		AuthPath:      authPath,
	}, nil
}

type ConfigMiddlewareGroup []string

// Path defines an individual API route (used in JSON/YAML)
type ConfigPath struct {
	Method      string `json:"method" yaml:"method"`
	Path        string `json:"path" yaml:"path"`
	ProxyTarget string `json:"proxy_target,omitempty" yaml:"proxy_target,omitempty"`
}

// ConfigRoute represents a route or proxy group from config
type ConfigRoute struct {
	Prefix          string       `json:"prefix" yaml:"prefix"`
	Method          string       `json:"method" yaml:"method"`
	Middleware      []string     `json:"middleware,omitempty" yaml:"middleware,omitempty"`
	MiddlewareGroup string       `json:"middleware_group,omitempty" yaml:"middleware_group,omitempty"`
	ProxyTarget     string       `json:"proxy_target,omitempty" yaml:"proxy_target"`
	Paths           []ConfigPath `json:"paths,omitempty" yaml:"paths,omitempty"`
}

// Config represents the full JSON/YAML structure
type InputConfig struct {
	MiddlewareGroups map[string]ConfigMiddlewareGroup `json:"middleware_groups" yaml:"middleware_groups"`
	Routes           []ConfigRoute                    `json:"routes" yaml:"routes"`
}

func LoadInputConfig(filepath string, fileType string) (*InputConfig, errors.ErrorHandler) {
	file, err := os.ReadFile(filepath)
	if err != nil {
		return nil, &errors.LoadConfigError{
			Message: fmt.Sprintf("error while reading config file: %v", err),
		}
	}

	var cfg InputConfig
	switch fileType {
	case "yaml":
		err = yaml.Unmarshal(file, &cfg)
	case "json":
		err = json.Unmarshal(file, &cfg)
	default:
		err = fmt.Errorf("unknown file type %s", fileType)
	}

	if err != nil {
		return nil, &errors.LoadConfigError{
			Message: fmt.Sprintf("error while unmarshaling config file: %v", err),
		}
	}

	return &cfg, nil
}
