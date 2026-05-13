package ports

import (
	"time"

	"github.com/google/uuid"
)

type MarkFailedPublishInput struct {
	EventID       uuid.UUID
	LastError     string
	LastAttemptAt *time.Time
	NextAttemptAt time.Time
}
