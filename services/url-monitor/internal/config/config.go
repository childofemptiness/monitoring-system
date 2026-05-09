package config

type Config struct {
	AppConfig           AppConfig
	RabbitMQConfig      RabbitMQConfig
	OutboxEventsConfig  OutboxEventsConfig
	MonitorChecksConfig MonitorChecksConfig
}
