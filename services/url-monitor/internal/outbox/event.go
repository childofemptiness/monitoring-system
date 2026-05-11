package outbox

import (
	"encoding/json"
	"time"

	"github.com/childofemptiness/monitoring-system/contracts/events"
	"github.com/google/uuid"
)

type Event struct {
	EventID             uuid.UUID
	EventType           events.EventType
	EventVersion        int
	Status              events.EventStatus
	Producer            events.EventProducer
	AttemptsCount       int
	LastError           error
	OccurredAt          time.Time
	CreatedAt           time.Time
	LastAttemptAt       *time.Time
	NextAttemptAt       time.Time
	ProcessingStartedAt *time.Time
	PublishedAt         *time.Time
	Payload             json.RawMessage
}
