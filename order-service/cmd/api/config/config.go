package config

import (
	"fmt"
	"github.com/joho/godotenv"
	constants "github.com/notoriouscode97/go-microservices/order-service/cmd/api/constants"
	"os"
	"strconv"
)

type Settings struct {
	Db struct {
		Dsn                string
		MaxOpenConnections int
		MaxIdleConnections int
		MaxIdleTime        string
	}
	Smtp struct {
		Host     string
		Port     int
		Username string
		Password string
		Sender   string
	}
	BindAddress       string
	CorsAllowedOrigin string
}

func newSettings(
	dsn string,
	maxOpenConnections int,
	maxIdleConnections int,
	maxIdleTime string,
	bindAddress string,
	corsAllowedOrigin string,
	smtpHost string,
	smtpPort int,
	smtpUsername string,
	smtpPassword string,
	smtpSender string,
) *Settings {
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
		Smtp: struct {
			Host     string
			Port     int
			Username string
			Password string
			Sender   string
		}{
			Host:     smtpHost,
			Port:     smtpPort,
			Username: smtpUsername,
			Password: smtpPassword,
			Sender:   smtpSender,
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

	smtpPort := getEnvOrPanic(constants.EnvKeys.SmtpPort)
	port, err := strconv.Atoi(smtpPort)

	if err != nil {
		port = 25
	}

	return newSettings(
		getEnvOrPanic(constants.EnvKeys.DatabaseDSN),
		25,
		25,
		"15m",
		getEnvOrPanic(constants.EnvKeys.ServerAddress),
		getEnvOrPanic(constants.EnvKeys.CorsAllowedOrigin),
		getEnvOrPanic(constants.EnvKeys.SmtpHost),
		port,
		getEnvOrPanic(constants.EnvKeys.SmtpUsername),
		getEnvOrPanic(constants.EnvKeys.SmtpPassword),
		getEnvOrPanic(constants.EnvKeys.SmtpSender),
	), nil
}

func getEnvOrPanic(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("environment variable %s not set", key))
	}
	return value
}
