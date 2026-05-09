package config

import "time"

const (
	monitorChecksLimit        = "MONITOR_CHECKS_LIMIT"
	monitorChecksTimeout      = "MONITOR_CHECKS_TIMEOUT"
	monitorChecksQueueSize    = "MONITOR_CHECKS_QUEUE_SIZE"
	monitorChecksWorkersCount = "MONITOR_CHECKS_WORKERS_COUNT"
)

type MonitorChecksConfig struct {
	MonitorChecksLimit       int
	MonitorCheckQueueSize    int
	MonitorSchedulerTimeout  time.Duration
	MonitorCheckWorkersCount int
}

func LoadMonitorChecksConfig() (MonitorChecksConfig, error) {
	workersCount, err := getEnvInt(monitorChecksWorkersCount, "5")
	if err != nil {
		return MonitorChecksConfig{}, err
	}

	queueSize, err := getEnvInt(monitorChecksQueueSize, "50")
	if err != nil {
		return MonitorChecksConfig{}, err
	}

	schedulerTimeout, err := getEnvDuration(monitorChecksTimeout, "2s")
	if err != nil {
		return MonitorChecksConfig{}, err
	}

	checksLimit, err := getEnvInt(monitorChecksLimit, "10")
	if err != nil {
		return MonitorChecksConfig{}, err
	}

	return MonitorChecksConfig{
		MonitorChecksLimit:       checksLimit,
		MonitorCheckQueueSize:    queueSize,
		MonitorSchedulerTimeout:  schedulerTimeout,
		MonitorCheckWorkersCount: workersCount,
	}, nil
}
