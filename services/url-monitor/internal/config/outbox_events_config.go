package config

import "time"

type OutboxEventsConfig struct {
	FetchInterval     time.Duration
	ProcessingTimeout time.Duration
	RetryBackoff      time.Duration
	MaxAttempts       int
}

func LoadOutboxEventsConfig() (OutboxEventsConfig, error) {
	fetchInterval, err := getEnvDuration("MONITOR_FETCH_INTERVAL", "2s")
	if err != nil {
		return OutboxEventsConfig{}, err
	}

	processingTimeout, err := getEnvDuration("MONITOR_PROCESSING_TIMEOUT", "2s")
	if err != nil {
		return OutboxEventsConfig{}, err
	}

	retryBackoff, err := getEnvDuration("MONITOR_RETRY_BACKOFF", "1s")
	if err != nil {
		return OutboxEventsConfig{}, err
	}

	maxAttempts, err := getEnvInt("MONITOR_MAX_ATTEMPTS", "5")
	if err != nil {
		return OutboxEventsConfig{}, err
	}

	return OutboxEventsConfig{
		FetchInterval:     fetchInterval,
		ProcessingTimeout: processingTimeout,
		RetryBackoff:      retryBackoff,
		MaxAttempts:       maxAttempts,
	}, nil
}
