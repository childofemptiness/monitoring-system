package rabbitmq_module

import (
	"context"
	"errors"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
)

type fakeChannel struct {
	declareCalled bool
	publishCalled bool
	closeCalled   bool

	declaredQueue string
	publishedMsg  amqp.Publishing

	gotCtx context.Context

	declareErr error
	publishErr error
	closeErr   error
}

func (c *fakeChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	c.declareCalled = true
	c.declaredQueue = name
	return amqp.Queue{}, c.declareErr
}

func (c *fakeChannel) PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	c.publishCalled = true
	c.gotCtx = ctx
	c.publishedMsg = msg

	if c.gotCtx.Err() != nil {
		return c.gotCtx.Err()
	}

	return c.publishErr
}

func (c *fakeChannel) Close() error {
	c.closeCalled = true
	return c.closeErr
}

func TestRabbitMQ_InvalidConfigReturnsError(t *testing.T) {
	client, err := NewRabbitMQClient(Config{})

	require.Error(t, err)
	require.Nil(t, client)
}

func TestRabbitMQ_ValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr error
	}{
		{
			name: "valid",
			cfg: Config{
				ConnectionURL: "amqp://guest:guest@localhost:5673/",
				QueueName:     "test-queue",
				PrefetchCount: 1,
				PrefetchSize:  0,
			},
			wantErr: nil,
		},
		{
			name: "empty connection url",
			cfg: Config{
				ConnectionURL: "",
				QueueName:     "test-queue",
				PrefetchCount: 1,
				PrefetchSize:  0,
			},
			wantErr: ErrEmptyConnectionURL,
		},
		{
			name: "invalid connection url",
			cfg: Config{
				ConnectionURL: "amqp://guest:guest@localhost:/",
				QueueName:     "test-queue",
				PrefetchCount: 1,
				PrefetchSize:  0,
			},
			wantErr: ErrInvalidConnectionURL,
		},
		{
			name: "empty queue name",
			cfg: Config{
				ConnectionURL: "amqp://guest:guest@localhost:5673/",
				QueueName:     "",
				PrefetchCount: 1,
				PrefetchSize:  0,
			},
			wantErr: ErrEmptyQueueName,
		},
		{
			name: "invalid prefetch count",
			cfg: Config{
				ConnectionURL: "amqp://guest:guest@localhost:5673/",
				QueueName:     "test-queue",
				PrefetchCount: 0,
				PrefetchSize:  0,
			},
			wantErr: ErrInvalidPrefetchCount,
		},
		{
			name: "invalid prefetch size",
			cfg: Config{
				ConnectionURL: "amqp://guest:guest@localhost:5673/",
				QueueName:     "test-queue",
				PrefetchCount: 1,
				PrefetchSize:  -1,
			},
			wantErr: ErrInvalidPrefetchSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.cfg)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestRabbitMQ_PublishMessageSuccess(t *testing.T) {
	cfg := Config{
		ConnectionURL: "amqp://guest:guest@localhost:5673/",
		QueueName:     "test.publish.queue",
		PrefetchCount: 1,
		PrefetchSize:  0,
	}
	fake := &fakeChannel{}
	client := &RabbitMQ{
		channel: fake,
		cfg:     cfg,
	}

	body := []byte(`{"event": "test"}`)

	err := client.Publish(context.Background(), body)

	require.NoError(t, err)
	require.Truef(t, fake.declareCalled, "declare was not called")
	require.Truef(t, fake.publishCalled, "publish was not called")
	require.Equal(t, cfg.QueueName, fake.declaredQueue)
	require.Equal(t, body, fake.publishedMsg.Body)
}

func TestRabbitMQ_PublishMessageCancelled(t *testing.T) {
	cfg := Config{
		ConnectionURL: "amqp://guest:guest@localhost:5673/",
		QueueName:     "test.publish.queue",
		PrefetchCount: 1,
		PrefetchSize:  0,
	}

	fake := &fakeChannel{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := &RabbitMQ{
		cfg:     cfg,
		channel: fake,
	}

	body := []byte(`{"event": "test"}`)
	err := client.Publish(ctx, body)
	require.Error(t, err)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("publish should have canceled the context")
	}
}

func TestRabbitMQ_PublishMessageTimeout(t *testing.T) {
	cfg := Config{
		ConnectionURL: "amqp://guest:guest@localhost:5673/",
		QueueName:     "test.publish.queue",
		PrefetchCount: 1,
		PrefetchSize:  0,
	}

	fake := &fakeChannel{}

	ctx, cancel := context.WithTimeout(context.Background(), -time.Second)
	defer cancel()

	client := &RabbitMQ{
		cfg:     cfg,
		channel: fake,
	}

	body := []byte(`{"event": "test"}`)

	err := client.Publish(ctx, body)
	require.Error(t, err)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("publish should have canceled the context")
	}
}
