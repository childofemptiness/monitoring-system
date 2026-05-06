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
