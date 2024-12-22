package config

import (
	"api_gateway/errors"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	AuthAddress       string
	AuthPageAddress   string
	ApiGatewayAddress string
	AuthPagePath      string
	AuthPath          string
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
		AuthAddress:       authAddress,
		AuthPageAddress:   authPageAddress,
		ApiGatewayAddress: apiGatewayAddress,
		AuthPagePath:      authPagePath,
		AuthPath:          authPath,
	}, nil
}
