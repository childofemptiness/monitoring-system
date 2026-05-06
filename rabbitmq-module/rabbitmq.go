package rabbitmq_module

import (
	"net/url"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	cfg     Config
}

func NewRabbitMQClient(cfg Config) (*RabbitMQ, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	conn, err := amqp.Dial(cfg.ConnectionURL)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()

		return nil, err
	}

	return &RabbitMQ{
		conn:    conn,
		channel: channel,
		cfg:     cfg,
	}, nil
}

func validateConfig(cfg Config) error {
	if cfg.ConnectionURL == "" {
		return ErrEmptyConnectionURL
	}

	raw, err := url.Parse(cfg.ConnectionURL)
	if err != nil {
		return ErrInvalidConnectionURL
	}

	if raw.Scheme != "amqp" && raw.Scheme != "amqps" {
		return ErrInvalidConnectionURL
	}

	if raw.Hostname() == "" || raw.Port() == "" {
		return ErrInvalidConnectionURL
	}

	if cfg.QueueName == "" {
		return ErrEmptyQueueName
	}

	if cfg.PrefetchCount <= 0 {
		return ErrInvalidPrefetchCount
	}

	if cfg.PrefetchSize < 0 {
		return ErrInvalidPrefetchSize
	}

	return nil
}
