package ports

import (
	"time"

	"github.com/google/uuid"
)

type MarkFailedPublishInput struct {
	eventID       uuid.UUID
	lastError     string
	lastAttemptAt time.Time
	nextAttemptAt *time.Time
}
