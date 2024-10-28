package config

import (
	"fmt"
	"github.com/joho/godotenv"
	constants "github.com/notoriouscode97/go-microservices/product-api/cmd/api/constants"
	"os"
)

type Settings struct {
	Db struct {
		Dsn                string
		MaxOpenConnections int
		MaxIdleConnections int
		MaxIdleTime        string
	}
	BindAddress       string
	CorsAllowedOrigin string
}

func newSettings(dsn string, maxOpenConnections int, maxIdleConnections int, maxIdleTime string, bindAddress string, corsAllowedOrigin string) *Settings {
	return &Settings{
		Db: struct {
			Dsn                string
			MaxOpenConnections int
			MaxIdleConnections int
			MaxIdleTime        string
		}{
			Dsn:                dsn,
			MaxOpenConnections: maxOpenConnections,
			MaxIdleConnections: maxIdleConnections,
			MaxIdleTime:        maxIdleTime,
		},
		BindAddress:       bindAddress,
		CorsAllowedOrigin: corsAllowedOrigin,
	}
}

func LoadConfig() (*Settings, error) {
	err := godotenv.Load("dev.env")

	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	return newSettings(
		getEnvOrPanic(constants.EnvKeys.DatabaseDSN),
		25,
		25,
		"15m",
		getEnvOrPanic(constants.EnvKeys.ServerAddress),
		getEnvOrPanic(constants.EnvKeys.CorsAllowedOrigin),
	), nil
}

func getEnvOrPanic(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("environment variable %s not set", key))
	}
	return value
}
