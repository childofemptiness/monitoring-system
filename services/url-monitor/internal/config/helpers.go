package config

import (
	"os"
	"strconv"
	"time"
)

func getEnvString(key, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}

func getEnvInt(key string, fallback string) (int, error) {
	value, err := strconv.Atoi(getEnvString(key, fallback))
	if err != nil {
		return value, err
	}

	return value, nil
}

func getEnvDuration(key string, fallback string) (time.Duration, error) {
	value, err := time.ParseDuration(getEnvString(key, fallback))
	if err != nil {
		return time.Duration(0), err
	}

	return value, nil
}
