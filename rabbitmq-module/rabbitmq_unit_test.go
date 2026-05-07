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
	declareCalled            bool
	publishWithContextCalled bool
	consumeWithContextCalled bool
	qosCalled                bool
	closeCalled              bool

	declaredQueue   string
	prefetchedCount int
	prefetchedSize  int
	publishedMsg    amqp.Publishing

	gotCtx context.Context

	declareErr            error
	publishWithContextErr error
	consumeWithContextErr error
	qosErr                error
	closeErr              error

	consumeDeliveries <-chan amqp.Delivery
	handledDelivery   rabbitDelivery
}

type fakeDelivery struct {
	ackCalled  bool
	nackCalled bool

	ackMultiple  bool
	nackMultiple bool
	nackRequeue  bool

	ackErr  error
	nackErr error

	body []byte
}

func (d *fakeDelivery) Body() []byte {
	return d.body
}

func (d *fakeDelivery) Ack(multiple bool) error {
	d.ackCalled = true
	d.ackMultiple = multiple
	return d.ackErr
}

func (d *fakeDelivery) Nack(multiple, requeue bool) error {
	d.nackCalled = true
	d.nackMultiple = multiple
	d.nackRequeue = requeue
	return d.nackErr
}

func (c *fakeChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	c.declareCalled = true
	c.declaredQueue = name
	return amqp.Queue{}, c.declareErr
}

func (c *fakeChannel) PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	c.publishWithContextCalled = true
	c.gotCtx = ctx
	c.publishedMsg = msg

	if c.gotCtx.Err() != nil {
		return c.gotCtx.Err()
	}

	return c.publishWithContextErr
}

func (c *fakeChannel) ConsumeWithContext(ctx context.Context, queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	c.consumeWithContextCalled = true
	c.gotCtx = ctx
	return c.consumeDeliveries, c.consumeWithContextErr
}

func (c *fakeChannel) Qos(prefetchCount int, prefetchSize int, global bool) error {
	c.qosCalled = true
	c.prefetchedCount = prefetchCount
	c.prefetchedSize = prefetchSize
	return c.qosErr
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

func TestRabbitMQ_PublishSuccess(t *testing.T) {
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
	ctx := context.Background()

	err := client.Publish(ctx, body)

	require.NoError(t, err)
	require.True(t, fake.declareCalled, "declare was not called")
	require.True(t, fake.publishWithContextCalled, "publish was not called")
	require.Equal(t, cfg.QueueName, fake.declaredQueue)
	require.Equal(t, ctx, fake.gotCtx)
	require.Equal(t, body, fake.publishedMsg.Body)
}

func TestRabbitMQ_PublishDeclareQueueError(t *testing.T) {
	expectedErr := errors.New("declare queue error")
	fake := &fakeChannel{}
	fake.declareErr = expectedErr

	client := &RabbitMQ{
		channel: fake,
	}

	err := client.Publish(context.Background(), []byte(`{"event": "test"}`))
	require.ErrorIs(t, expectedErr, err)
	require.True(t, fake.declareCalled, "declare was not called")
}

func TestRabbitMQ_PublishCancelled(t *testing.T) {
	fake := &fakeChannel{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := &RabbitMQ{
		cfg:     testConfig(),
		channel: fake,
	}

	body := []byte(`{"event": "test"}`)
	err := client.Publish(ctx, body)
	require.Error(t, err, context.Canceled, "publish should have canceled the context")
}

func TestRabbitMQ_PublishTimeout(t *testing.T) {
	fake := &fakeChannel{}

	ctx, cancel := context.WithTimeout(context.Background(), -time.Second)
	defer cancel()

	client := &RabbitMQ{
		cfg:     testConfig(),
		channel: fake,
	}

	body := []byte(`{"event": "test"}`)

	err := client.Publish(ctx, body)
	require.Error(t, err, context.DeadlineExceeded, "publish should have canceled the context")
}

func TestRabbitMQ_PublishPublishWithContextError(t *testing.T) {
	expectedErr := errors.New("publish with context error")
	fake := &fakeChannel{}
	fake.publishWithContextErr = expectedErr

	client := &RabbitMQ{
		channel: fake,
	}

	err := client.Publish(context.Background(), []byte(`{"event": "test"}`))
	require.ErrorIs(t, expectedErr, err)
	require.True(t, fake.declareCalled, "declare was not called")
	require.True(t, fake.publishWithContextCalled, "publish was not called")
}

func TestRabbitMQ_ConsumeQueueDeclareError(t *testing.T) {
	fake := &fakeChannel{}
	expectedErr := errors.New("queue declare failed")
	fake.declareErr = expectedErr

	client := &RabbitMQ{
		cfg:     testConfig(),
		channel: fake,
	}

	err := client.Consume(context.Background(), func(ctx context.Context, body []byte) error {
		return nil
	})
	require.ErrorIs(t, err, expectedErr)
	require.True(t, fake.declareCalled, "declare was not called")
	require.False(t, fake.qosCalled, "qos was not called")
	require.False(t, fake.consumeWithContextCalled, "consume shouldn't have been called")
}

func TestRabbitMQ_ConsumeQosError(t *testing.T) {
	fake := &fakeChannel{}
	expectedErr := errors.New("qos call failed")
	fake.qosErr = expectedErr

	client := &RabbitMQ{
		cfg:     testConfig(),
		channel: fake,
	}

	err := client.Consume(context.Background(), func(ctx context.Context, body []byte) error {
		return nil
	})
	require.ErrorIs(t, err, expectedErr)
	require.True(t, fake.declareCalled, "declare was not called")
	require.True(t, fake.qosCalled, "qos was not called")
	require.False(t, fake.consumeWithContextCalled, "consume shouldn't have been called")
	require.Equal(t, client.cfg.QueueName, fake.declaredQueue)
}

func TestRabbitMQ_ConsumeConsumeWithContextError(t *testing.T) {
	fake := &fakeChannel{}
	expectedErr := errors.New("consume call failed")
	fake.consumeWithContextErr = expectedErr

	client := &RabbitMQ{
		cfg:     testConfig(),
		channel: fake,
	}

	err := client.Consume(context.Background(), func(ctx context.Context, body []byte) error {
		return nil
	})
	require.ErrorIs(t, err, expectedErr)
	require.True(t, fake.declareCalled, "declare was not called")
	require.True(t, fake.qosCalled, "qos was not called")
	require.True(t, fake.consumeWithContextCalled, "consume was not called")
	require.Equal(t, client.cfg.QueueName, fake.declaredQueue)
	require.Equal(t, client.cfg.PrefetchCount, fake.prefetchedCount)
	require.Equal(t, client.cfg.PrefetchSize, fake.prefetchedSize)
}

func TestRabbitMQ_HandleMessageNackError(t *testing.T) {
	handleErr := errors.New("handle call failed")
	expectedErr := errors.New("message nack failed")

	client := &RabbitMQ{}
	fake := fakeDelivery{
		body: []byte(`{"event": "test"}`),
	}
	fake.nackErr = expectedErr

	err := client.handleMessage(context.Background(), &fake, func(ctx context.Context, body []byte) error {
		return handleErr
	})
	require.ErrorIs(t, err, expectedErr)
	require.True(t, fake.nackCalled, "Nack was not called")
	require.False(t, fake.ackCalled, "Ack shouldn't have been called")
}

func TestRabbitMQ_HandleMessageNackSuccess(t *testing.T) {
	handlerErr := errors.New("handle call failed")

	client := &RabbitMQ{}
	fake := &fakeDelivery{
		body: []byte(`{"event": "test"}`),
	}

	err := client.handleMessage(context.Background(), fake, func(ctx context.Context, body []byte) error {
		return handlerErr
	})
	require.Nil(t, err)
	require.True(t, fake.nackCalled, "Nack was not called")
	require.False(t, fake.nackMultiple, "Nack multiple wasn't false")
	require.True(t, fake.nackRequeue, "Nack requeue wasn't true")
	require.False(t, fake.ackCalled, "Ack shouldn't have been called")
}

func TestRabbitMQ_HandleMessageErrNonRetryable(t *testing.T) {
	client := &RabbitMQ{}
	fake := &fakeDelivery{
		body: []byte(`{"event": "test"}`),
	}

	err := client.handleMessage(context.Background(), fake, func(ctx context.Context, body []byte) error {
		return ErrNonRetryable
	})
	require.Nil(t, err)
	require.True(t, fake.nackCalled, "Nack was not called")
	require.False(t, fake.nackMultiple, "Nack multiple wasn't false")
	require.False(t, fake.nackRequeue, "Nack requeue wasn't false")
	require.False(t, fake.ackCalled, "Ack shouldn't have been called")
}

func TestRabbitMQ_HandleMessageAckSuccess(t *testing.T) {
	fake := &fakeDelivery{
		body: []byte(`{"event": "test"}`),
	}
	client := &RabbitMQ{}

	err := client.handleMessage(context.Background(), fake, func(ctx context.Context, body []byte) error {
		return nil
	})
	require.Nil(t, err)
	require.False(t, fake.nackCalled, "Nack shouldn't have been called")
	require.True(t, fake.ackCalled, "Ack was not called")
	require.False(t, fake.ackMultiple, "Ack multiple wasn't false")
}

func TestRabbitMQ_HandleMessageAckError(t *testing.T) {
	client := &RabbitMQ{}
	fake := fakeDelivery{
		body: []byte(`{"event": "test"}`),
	}
	fake.ackErr = errors.New("ack failed")

	err := client.handleMessage(context.Background(), &fake, func(ctx context.Context, body []byte) error {
		return nil
	})
	require.ErrorIs(t, err, fake.ackErr)
	require.False(t, fake.nackCalled, "Nack shouldn't have been called")
	require.True(t, fake.ackCalled, "ack shouldn't have been called")
}

func testConfig() Config {
	return Config{
		ConnectionURL: "amqp://guest:guest@localhost:5673/",
		QueueName:     "test.queue",
		PrefetchCount: 1,
		PrefetchSize:  0,
	}
}
