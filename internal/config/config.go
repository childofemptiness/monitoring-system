package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppPort 	string
	DatabaseURL string
}

func Load() (Config, error) {
	cfg := Config{
		AppPort: getEnv("APP_PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	
	if value == "" {
		return fallback
	}

	return value
}
