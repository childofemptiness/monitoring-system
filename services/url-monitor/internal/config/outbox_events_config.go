package config

import "time"

const (
	maxAttempts       = "OUTBOX_MAX_ATTEMPTS"
	retryBackoff      = "OUTBOX_RETRY_BACKOFF"
	fetchInterval     = "OUTBOX_FETCH_INTERVAL"
	processingTimeout = "OUTBOX_PROCESSING_TIMEOUT"
	fetchEventsLimit  = "OUTBOX_EVENTS_LIMIT"
	workersCount      = "OUTBOX_WORKERS_COUNT"
	queueSize         = "OUTBOX_QUEUE_SIZE"
)

type OutboxEventsConfig struct {
	MaxAttempts       int
	RetryBackoff      time.Duration
	FetchInterval     time.Duration
	FetchEventsLimit  int
	ProcessingTimeout time.Duration
	WorkersCount      int
	QueueSize         int
}

func LoadOutboxEventsConfig() (OutboxEventsConfig, error) {
	fetchInterval, err := getEnvDuration(fetchInterval, "2s")
	if err != nil {
		return OutboxEventsConfig{}, err
	}

	processingTimeout, err := getEnvDuration(processingTimeout, "2s")
	if err != nil {
		return OutboxEventsConfig{}, err
	}

	retryBackoff, err := getEnvDuration(retryBackoff, "1s")
	if err != nil {
		return OutboxEventsConfig{}, err
	}

	fetchEventsLimit, err := getEnvInt(fetchEventsLimit, "10")
	if err != nil {
		return OutboxEventsConfig{}, err
	}

	workersCount, err := getEnvInt(workersCount, "10")
	if err != nil {
		return OutboxEventsConfig{}, err
	}

	queueSize, err := getEnvInt(queueSize, "10")
	if err != nil {
		return OutboxEventsConfig{}, err
	}

	maxAttempts, err := getEnvInt(maxAttempts, "5")
	if err != nil {
		return OutboxEventsConfig{}, err
	}

	return OutboxEventsConfig{
		FetchInterval:     fetchInterval,
		ProcessingTimeout: processingTimeout,
		RetryBackoff:      retryBackoff,
		FetchEventsLimit:  fetchEventsLimit,
		MaxAttempts:       maxAttempts,
		WorkersCount:      workersCount,
		QueueSize:         queueSize,
	}, nil
}
