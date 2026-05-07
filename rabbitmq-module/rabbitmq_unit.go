package rabbitmq_module

import (
	"context"
	"errors"
	"net/url"

	amqp "github.com/rabbitmq/amqp091-go"
)

type amqpChannel interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	PublishWithContext(ctx context.Context, exchange, key string, mandatory bool, immediate bool, msg amqp.Publishing) error
	ConsumeWithContext(ctx context.Context, queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	Qos(prefetchCount, prefetchSize int, global bool) error
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

func (r *RabbitMQ) Consume(ctx context.Context, handler func(ctx context.Context, body []byte) error) error {
	queue, err := r.declareQueue()
	if err != nil {
		return err
	}

	if err := r.channel.Qos(
		r.cfg.PrefetchCount,
		r.cfg.PrefetchSize,
		false,
	); err != nil {
		return err
	}

	deliveries, err := r.channel.ConsumeWithContext(
		ctx,
		queue.Name,
		r.cfg.ConsumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case delivery, ok := <-deliveries:
			if !ok {
				return nil
			}

			if err := r.handleMessage(ctx, &rabbitDelivery{msg: delivery}, handler); err != nil {
				return err
			}
		}
	}
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

func (r *RabbitMQ) handleMessage(
	ctx context.Context,
	delivery acknowledger,
	handler func(ctx context.Context, body []byte) error,
) error {
	// TODO: configure DLQ before using non-retryable rejection in production.
	if err := handler(ctx, delivery.Body()); err != nil {
		requeue := true

		if errors.Is(err, ErrNonRetryable) {
			requeue = false
		}

		if err = delivery.Nack(false, requeue); err != nil {
			return err
		}

		return nil
	}

	if err := delivery.Ack(false); err != nil {
		return err
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
