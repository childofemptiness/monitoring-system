package config

import "time"

const (
	monitorChecksWorkersCountPropertyName = "MONITOR_CHECKS_WORKERS_COUNT"
	monitorChecksQueueSizePropertyName    = "MONITOR_CHECKS_QUEUE_SIZE"
	monitorChecksTimeoutPropertyName      = "MONITOR_CHECKS_TIMEOUT"
)

type MonitorChecksConfig struct {
	MonitorCheckQueueSize    int
	MonitorCheckWorkersCount int
	MonitorSchedulerTimeout  time.Duration
}

func LoadMonitorChecksConfig() (MonitorChecksConfig, error) {
	workersCount, err := getEnvInt(monitorChecksWorkersCountPropertyName, "5")
	if err != nil {
		return MonitorChecksConfig{}, err
	}

	queueSize, err := getEnvInt(monitorChecksQueueSizePropertyName, "50")
	if err != nil {
		return MonitorChecksConfig{}, err
	}

	schedulerTimeout, err := getEnvDuration(monitorChecksTimeoutPropertyName, "2s")
	if err != nil {
		return MonitorChecksConfig{}, err
	}

	return MonitorChecksConfig{
		MonitorCheckQueueSize:    queueSize,
		MonitorCheckWorkersCount: workersCount,
		MonitorSchedulerTimeout:  schedulerTimeout,
	}, nil
}
