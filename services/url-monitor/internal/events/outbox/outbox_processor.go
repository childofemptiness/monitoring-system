package outbox

import (
	"context"
	"time"
	"url-monitor/internal/events"
	"url-monitor/internal/ports"

	"github.com/google/uuid"
)

type OutboxService interface {
	MarkForRetry(ctx context.Context, input ports.MarkFailedPublishInput) error
	MarkExhausted(ctx context.Context, input ports.MarkFailedPublishInput) error
	MarkPublished(ctx context.Context, eventID uuid.UUID, publishedAt time.Time) error
}

type EventPublisher interface {
	Publish(ctx context.Context, event events.EventEnvelope) error
}

type EventProcessor struct {
	service   OutboxService
	publisher EventPublisher
}
