package config

import (
	"fmt"
	"os"
)

const (
	appPortPropertyName     = "APP_PORT"
	databaseURLPropertyName = "DATABASE_URL"
)

type AppConfig struct {
	AppPort     string
	DatabaseURL string
}

func LoadAppConfig() (AppConfig, error) {
	cfg := AppConfig{
		AppPort:     getEnvString(appPortPropertyName, "8080"),
		DatabaseURL: os.Getenv(databaseURLPropertyName),
	}

	if cfg.DatabaseURL == "" {
		return AppConfig{}, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}
