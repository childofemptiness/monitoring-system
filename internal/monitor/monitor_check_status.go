package monitor

type MonitorCheckStatus string

const (
	MonitorCheckStatusUp    MonitorCheckStatus = "up"
	MonitorCheckStatusDown  MonitorCheckStatus = "down"
	MonitorCheckStatusError MonitorCheckStatus = "error"
)
