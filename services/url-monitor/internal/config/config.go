package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppPort                  string
	DatabaseURL              string
	MonitorCheckWorkersCount int
	MonitorCheckQueueSize    int
	SchedulerTimeInterval    int
}

func Load() (Config, error) {
	workersCount, err := getEnvInt("MONITOR_CHECKS_WORKERS_COUNT", "5")
	if err != nil {
		return Config{}, err
	}

	queueSize, err := getEnvInt("MONITOR_CHECKS_QUEUE_SIZE", "50")
	if err != nil {
		return Config{}, err
	}

	schedulerTimeInterval, err := getEnvInt("SCHEDULER_TIME_INTERVAL", "2")
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		AppPort:                  getEnvString("APP_PORT", "8080"),
		DatabaseURL:              os.Getenv("DATABASE_URL"),
		MonitorCheckWorkersCount: workersCount,
		MonitorCheckQueueSize:    queueSize,
		SchedulerTimeInterval:    schedulerTimeInterval,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

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
