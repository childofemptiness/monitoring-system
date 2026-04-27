package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type EventEnvelope struct {
	EventID      uuid.UUID
	EventType    EventType
	EventVersion int
	OccurredAt   time.Time
	Producer     EventProducer
	Payload      json.RawMessage
}
