package outbox

import (
	"context"
	"time"

	"github.com/childofemptiness/url-monitor/internal/ports"
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

func (o OutboxEventService) MarkForRetry(ctx context.Context, input ports.MarkFailedPublishInput) error {
	//TODO implement me
	return nil
}

func (o OutboxEventService) MarkExhausted(ctx context.Context, input ports.MarkFailedPublishInput) error {
	//TODO implement me
	return nil
}

func (o OutboxEventService) MarkPublished(ctx context.Context, eventID uuid.UUID, publishedAt time.Time) error {
	//TODO implement me
	return nil
}

func NewOutboxEventService(repo MarkEventRepository) *OutboxEventService {
	return &OutboxEventService{repo: repo}
}
