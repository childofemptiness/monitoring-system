package outbox

import (
	"time"
	"url-monitor/internal/monitor"
)

const EventVersionURLChecked = 1

type URLCheckedPayload struct {
	CheckID        int64                      `json:"check_id"`
	MonitorID      int64                      `json:"monitor_id"`
	URL            string                     `json:"url"`
	Status         monitor.MonitorCheckStatus `json:"status"`
	HTTPStatusCode *int16                     `json:"http_status_code,omitempty"`
	ErrorKind      *monitor.CheckErrorKind    `json:"error_kind,omitempty"`
	CheckedAt      time.Time                  `json:"checked_at"`
}
