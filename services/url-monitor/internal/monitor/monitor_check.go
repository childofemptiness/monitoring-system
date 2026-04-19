package monitor

import (
	"time"
)

type MonitorCheck struct {
	ID             int64
	MonitorID      int64
	Status         MonitorCheckStatus
	HTTPStatusCode int16
	ErrorKind      CheckErrorKind
	ErrorMessage   string
	ResponseTimeMS int64
	StartedAt      time.Time
	FinishedAt     time.Time
}
