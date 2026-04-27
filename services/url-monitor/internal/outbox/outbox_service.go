package outbox

import (
	"context"
	"time"
	"url-monitor/internal/ports"

	"github.com/google/uuid"
)

type MarkEventRepository interface {
	MarkForRetry(ctx context.Context, input ports.MarkFailedPublishInput) error
	MarkExhausted(ctx context.Context, input ports.MarkFailedPublishInput) error
	MarkPublished(ctx context.Context, publishedAt time.Time, eventID uuid.UUID) error
}

type OutboxEventService struct {
	repo MarkEventRepository
}
