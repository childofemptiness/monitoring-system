package outbox

import (
	"context"
	"time"

	"github.com/childofemptiness/url-monitor/internal/config"
	"github.com/childofemptiness/url-monitor/internal/ports"
	"github.com/childofemptiness/url-monitor/internal/scheduler"
)

type ClaimOutboxRepository interface {
	ClaimNextBatch(ctx context.Context, input ports.ClaimNextBatchInput) ([]Event, error)
}

type EventDispatcher interface {
	Submit(ctx context.Context, event Event) error
}

type EventFetcher struct {
	repo          ClaimOutboxRepository
	dispatcher    EventDispatcher
	cfg           config.OutboxEventsConfig
	baseScheduler scheduler.Scheduler[Event]
}

func NewEventFetcher(repo ClaimOutboxRepository, dispatcher EventDispatcher, cfg config.OutboxEventsConfig) *EventFetcher {
	return &EventFetcher{
		repo:       repo,
		dispatcher: dispatcher,
		cfg:        cfg,
	}
}

func (f *EventFetcher) Run(ctx context.Context) error {
	if err := f.baseScheduler.Run(
		ctx,
		f.cfg.FetchInterval,
		func(ctx context.Context) ([]Event, error) {
			return f.repo.ClaimNextBatch(ctx, ports.ClaimNextBatchInput{
				FetchEventsLimit:  f.cfg.FetchEventsLimit,
				ProcessingTimeout: f.cfg.ProcessingTimeout,
				Now:               time.Now(),
			})
		},
		f.dispatcher.Submit,
	); err != nil {
		return err
	}

	return nil
}
