package config

import rabbitmqmodule "github.com/childofemptiness/rabbitmq-module"

const (
	rabbitQueueName       = "RABBITMQ_QUEUE_NAME"
	rabbitmqPrefetchSize  = "RABBITMQ_PREFETCH_SIZE"
	rabbitmqPrefetchCount = "RABBITMQ_PREFETCH_COUNT"
	rabbitmqConnectionURL = "RABBITMQ_CONNECTION_URL"
)

type RabbitMQConfig struct {
	QueueName     string
	PrefetchSize  int
	PrefetchCount int
	ConnectionURL string
}

func LoadRabbitMQConfig() (RabbitMQConfig, error) {
	prefetchCount, err := getEnvInt(rabbitmqPrefetchCount, "1")
	if err != nil {
		return RabbitMQConfig{}, err
	}

	prefetchSize, err := getEnvInt(rabbitmqPrefetchSize, "0")
	if err != nil {
		return RabbitMQConfig{}, err
	}

	return RabbitMQConfig{
		ConnectionURL: getEnvString(rabbitmqConnectionURL, ""),
		QueueName:     getEnvString(rabbitQueueName, ""),
		PrefetchCount: prefetchCount,
		PrefetchSize:  prefetchSize,
	}, nil
}

func (cfg RabbitMQConfig) Wrap() rabbitmqmodule.Config {
	return rabbitmqmodule.Config{
		ConnectionURL: cfg.ConnectionURL,
		QueueName:     cfg.QueueName,
		PrefetchCount: cfg.PrefetchCount,
		PrefetchSize:  cfg.PrefetchSize,
	}
}
