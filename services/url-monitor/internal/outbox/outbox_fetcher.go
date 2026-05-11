package outbox

import (
	"context"
	"time"

	"github.com/childofemptiness/url-monitor/internal/config"
	"github.com/childofemptiness/url-monitor/internal/scheduler"
)

type ClaimOutboxRepository interface {
	ClaimNextBatch(ctx context.Context, now time.Time, limit int) ([]Event, error)
}

type EventDispatcher interface {
	Dispatch(ctx context.Context, event Event) error
}

type EventFetcher struct {
	repo          ClaimOutboxRepository
	dispatcher    EventDispatcher
	cfg           config.OutboxEventsConfig
	baseScheduler scheduler.Scheduler[Event]
}

func (f *EventFetcher) Run(ctx context.Context) error {
	if err := f.baseScheduler.Run(
		ctx,
		f.cfg.FetchInterval,
		func(ctx context.Context) ([]Event, error) {
			return f.repo.ClaimNextBatch(ctx, time.Now(), f.cfg.FetchEventsLimit)
		},
		f.dispatcher.Dispatch,
	); err != nil {
		return err
	}

	return nil
}
