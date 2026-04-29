package rabbitmq_module

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRabbitMQ_InvalidConfigReturnsError(t *testing.T) {
	client, err := NewRabbitMQClient(Config{})

	require.Error(t, err)
	require.Nil(t, client)
}

func TestRabbitMQ_ValidateConfigSuccess(t *testing.T) {
	cfg := Config{
		ConnectionURL: "amqp://guest:guest@localhost:5672/",
		QueueName:     "test-queue",
		PrefetchCount: 10,
		PrefetchSize:  5,
		ConsumerTag:   "test-consumer",
	}

	if err := validateConfig(cfg); err != nil {
		t.Fatalf("config validation failed: %s", err)
	}
}
func TestRabbitMQ_ValidateConfigFailure(t *testing.T) {
	cfg := Config{
		ConnectionURL: "amqp://guest:guest@localhost:5672/",
		QueueName:     "test-queue",
		PrefetchCount: 0,
		PrefetchSize:  0,
		ConsumerTag:   "test-consumer",
	}

	if err := validateConfig(cfg); err == nil {
		t.Fatalf("config validation should fail")
	}
}
