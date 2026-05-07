package rabbitmq_module

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
)

func TestRabbitmq_ConnectionSuccessful(t *testing.T) {
	client, err := NewRabbitMQClient(integrationTestConfig(t))
	require.NoError(t, err)

	defer func(client *RabbitMQ) {
		err := client.Close()
		require.NoError(t, err)
	}(client)
}

func TestRabbitMQ_CloseErrorWhenConnClosed(t *testing.T) {
	client, err := NewRabbitMQClient(integrationTestConfig(t))
	require.NoError(t, err)

	err = client.Close()
	require.NoError(t, err)

	err = client.Close()
	require.ErrorIs(t, err, amqp091.ErrClosed)
}

func TestRabbitMQ_PublishSuccessful(t *testing.T) {
	client, err := NewRabbitMQClient(integrationTestConfig(t))
	require.NoError(t, err)

	defer func(client *RabbitMQ) {
		err := client.Close()
		require.NoError(t, err)
	}(client)

	err = client.Publish(context.Background(), []byte(`{"event": "test"}`))
	require.NoError(t, err)
}

func TestRabbitMQ_PublishCancelledContext(t *testing.T) {
	client, err := NewRabbitMQClient(integrationTestConfig(t))
	require.NoError(t, err)

	defer func(client *RabbitMQ) {
		err := client.Close()
		require.NoError(t, err)
	}(client)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = client.Publish(ctx, []byte(`{"event": "test"}`))
	require.ErrorIs(t, err, context.Canceled)
}

func TestRabbitMQ_Publish_Timeout(t *testing.T) {
	client, err := NewRabbitMQClient(integrationTestConfig(t))
	require.NoError(t, err)

	defer func(client *RabbitMQ) {
		err := client.Close()
		require.NoError(t, err)
	}(client)

	ctx, cancel := context.WithTimeout(context.Background(), -time.Second)
	defer cancel()

	err = client.Publish(ctx, []byte(`{"event": "test"}`))
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestRabbitMQ_ConsumeSuccessful(t *testing.T) {
	client, err := NewRabbitMQClient(integrationTestConfig(t))
	require.NoError(t, err)

	defer func(client *RabbitMQ) {
		err := client.Close()
		require.NoError(t, err)
	}(client)

	body := []byte(`{"event": "test"}`)
	ctx, cancel := context.WithCancel(context.Background())

	err = client.Publish(ctx, body)
	require.NoError(t, err)

	received := make(chan []byte, 1)
	consumeDone := make(chan error, 1)

	go func() {
		consumeDone <- client.Consume(ctx, func(ctx context.Context, body []byte) error {
			received <- body
			cancel()
			return nil
		})
	}()
	require.ErrorIs(t, <-consumeDone, context.Canceled)
	require.Equal(t, body, <-received)
}

func integrationTestConfig(t *testing.T) Config {
	t.Helper()

	err := godotenv.Load()
	if err != nil {
		t.Skip(".env file not found")
	}

	connectionURL := os.Getenv("RABBITMQ_TEST_URL")
	if connectionURL == "" {
		t.Skip("RabbitMQ integration tests skipped")
	}

	queueName := os.Getenv("RABBITMQ_TEST_QUEUE")
	if queueName == "" {
		t.Skip("RabbitMQ integration tests skipped")
	}

	return Config{
		ConnectionURL: connectionURL,
		QueueName:     uniqQueueName(t, queueName),
		PrefetchCount: 1,
		PrefetchSize:  0,
	}
}

func uniqQueueName(t *testing.T, queueName string) string {
	return queueName + "." + t.Name() + "." + strconv.FormatInt(time.Now().UnixNano(), 10)
}
