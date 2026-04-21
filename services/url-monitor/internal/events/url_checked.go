package events

import (
	"time"
	"url-monitor/internal/monitor"

	"github.com/google/uuid"
)

const EventVersionURLChecked = 1

type Payload struct {
	CheckID        int64                      `json:"check_id"`
	MonitorID      int64                      `json:"monitor_id"`
	URL            string                     `json:"url"`
	Status         monitor.MonitorCheckStatus `json:"status"`
	HTTPStatusCode *int16                     `json:"http_status_code,omitempty"`
	ErrorKind      *monitor.CheckErrorKind    `json:"error_kind,omitempty"`
	CheckedAt      time.Time                  `json:"checked_at"`
}

type URLChecked struct {
	EventID      uuid.UUID
	EventType    EventType
	EventVersion int
	OccurredAt   time.Time
	Producer     EventProducer
	Payload      Payload
}
