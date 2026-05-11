package outbox

import (
	"encoding/json"
	"time"

	"github.com/childofemptiness/monitoring-system/contracts/events"
	"github.com/google/uuid"
)

type Event struct {
	EventID      uuid.UUID
	EventType    events.EventType
	EventVersion int
	OccurredAt   time.Time
	Producer     events.EventProducer
	Payload      json.RawMessage
}
