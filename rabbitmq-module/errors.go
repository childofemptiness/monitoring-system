package rabbitmq_module

import "errors"

var (
	ErrEmptyConnectionURL   = errors.New("connection URL is empty")
	ErrInvalidConnectionURL = errors.New("invalid connection URL")
	ErrEmptyQueueName       = errors.New("queue name is empty")
	ErrInvalidPrefetchCount = errors.New("invalid prefetch count")
	ErrInvalidPrefetchSize  = errors.New("invalid prefetch size")
	ErrNonRetryable         = errors.New("non-retriable error")
)
