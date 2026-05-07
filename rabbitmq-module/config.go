package rabbitmq_module

type Config struct {
	QueueName     string
	PrefetchCount int
	PrefetchSize  int
	ConsumerTag   string
	ConnectionURL string
}
