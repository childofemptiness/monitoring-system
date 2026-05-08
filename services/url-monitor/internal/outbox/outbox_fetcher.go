package outbox

import (
	"context"
	"time"

	"github.com/childofemptiness/monitoring-system/contracts/events"
	"github.com/childofemptiness/url-monitor/internal/config"
	"github.com/childofemptiness/url-monitor/internal/scheduler"
)

type ClaimOutboxRepository interface {
	ClaimNextBatch(ctx context.Context, now time.Time, limit int) ([]events.EventEnvelope, error)
}

type EventDispatcher interface {
	Dispatch(ctx context.Context, event events.EventEnvelope) error
}

type EventFetcher struct {
	repo          ClaimOutboxRepository
	dispatcher    EventDispatcher
	cfg           config.OutboxEventsConfig
	baseScheduler scheduler.Scheduler[events.EventEnvelope]
}

func (f *EventFetcher) Run(ctx context.Context) error {
	if err := f.baseScheduler.Run(
		ctx,
		f.cfg.FetchInterval,
		func(ctx context.Context) ([]events.EventEnvelope, error) {
			return f.repo.ClaimNextBatch(ctx, time.Now(), f.cfg.FetchEventsLimit)
		},
		f.dispatcher.Dispatch,
	); err != nil {
		return err
	}

	return nil
}
