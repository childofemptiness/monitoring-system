package ports

import (
	"time"
	"url-monitor/internal/monitor"

	"github.com/childofemptiness/monitoring-system/contracts/events"
	"github.com/google/uuid"
)

type CreateCheckWithEventInput struct {
	MonitorID      int64
	URL            string
	Status         monitor.MonitorCheckStatus
	HTTPStatusCode int16
	ErrorMessage   string
	ErrorKind      monitor.CheckErrorKind
	ResponseTimeMS int64
	StartedAt      time.Time
	FinishedAt     time.Time

	EventID      uuid.UUID
	EventType    events.EventType
	EventVersion int
	OccurredAt   time.Time
	Producer     events.EventProducer

	NextCheckAt time.Time
}
