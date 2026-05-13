package outbox

import (
	"context"
	"time"

	"github.com/childofemptiness/url-monitor/internal/config"
	"github.com/childofemptiness/url-monitor/internal/ports"

	"github.com/childofemptiness/monitoring-system/contracts/events"
	"github.com/google/uuid"
)

type EventService interface {
	MarkForRetry(ctx context.Context, input ports.MarkFailedPublishInput) error
	MarkExhausted(ctx context.Context, input ports.MarkFailedPublishInput) error
	MarkPublished(ctx context.Context, eventID uuid.UUID, publishedAt time.Time) error
}

type EventPublisher interface {
	Publish(ctx context.Context, event events.EventEnvelope) error
}

type EventProcessor struct {
	cfg       config.OutboxEventsConfig
	service   EventService
	publisher EventPublisher
}

func NewEventProcessor(
	cfg config.OutboxEventsConfig,
	service EventService,
	publisher EventPublisher,
) *EventProcessor {
	return &EventProcessor{
		cfg:       cfg,
		service:   service,
		publisher: publisher,
	}
}

func (p *EventProcessor) Process(ctx context.Context, event Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	eventEnvelope := events.EventEnvelope{
		EventID:      event.EventID,
		EventType:    event.EventType,
		EventVersion: event.EventVersion,
		OccurredAt:   event.OccurredAt,
		Producer:     event.Producer,
		Payload:      event.Payload,
	}
	if err := p.publisher.Publish(ctx, eventEnvelope); err == nil {
		return p.service.MarkPublished(ctx, event.EventID, time.Now())
	} else {
		attemptsCount := event.AttemptsCount + 1
		lastAttemptAt := time.Now()

		input := ports.MarkFailedPublishInput{
			EventID:       event.EventID,
			LastError:     err.Error(),
			LastAttemptAt: &lastAttemptAt,
			NextAttemptAt: event.NextAttemptAt,
		}

		if attemptsCount >= p.cfg.MaxAttempts {
			return p.service.MarkExhausted(ctx, input)
		}

		input.NextAttemptAt = lastAttemptAt.Add(time.Duration(event.AttemptsCount) * p.cfg.RetryBackoff)

		return p.service.MarkForRetry(ctx, input)
	}
}
