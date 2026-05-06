package rabbitmq_module

import (
	"context"
	"net/url"

	amqp "github.com/rabbitmq/amqp091-go"
)

type amqpChannel interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	PublishWithContext(ctx context.Context, exchange, key string, mandatory bool, immediate bool, msg amqp.Publishing) error
	Close() error
}

type RabbitMQ struct {
	conn    *amqp.Connection
	channel amqpChannel
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

func (r *RabbitMQ) Publish(ctx context.Context, body []byte) error {
	if _, err := r.declareQueue(); err != nil {
		return err
	}

	return r.channel.PublishWithContext(
		ctx,
		"",
		r.cfg.QueueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		})
}

func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			return err
		}
	}

	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (r *RabbitMQ) declareQueue() (amqp.Queue, error) {
	return r.channel.QueueDeclare(
		r.cfg.QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
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
