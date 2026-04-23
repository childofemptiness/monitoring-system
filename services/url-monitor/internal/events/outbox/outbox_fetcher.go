package outbox

import (
	"context"
	"time"
	"url-monitor/internal/events"
)

type ClaimOutboxRepository interface {
	ClaimNextBatch(ctx context.Context, now time.Time, limit int) ([]events.EventEnvelope, error)
}

type EventDispatcher interface {
	Dispatch(ctx context.Context, event events.EventEnvelope) error
}

type EventFetcher struct {
	repo       ClaimOutboxRepository
	dispatcher EventDispatcher
}
