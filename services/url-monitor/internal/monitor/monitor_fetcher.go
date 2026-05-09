package monitor

import (
	"context"
	"time"

	"github.com/childofemptiness/url-monitor/internal/config"
	"github.com/childofemptiness/url-monitor/internal/scheduler"
)

type MonitorRepository interface {
	ListDue(ctx context.Context, now time.Time, limit int) ([]Monitor, error)
}

type Dispatcher interface {
	Submit(ctx context.Context, m Monitor) error
}

type MonitorFetcher struct {
	baseScheduler scheduler.Scheduler[Monitor]
	cfg           config.MonitorChecksConfig
	repo          MonitorRepository
	dispatcher    Dispatcher
}

func NewMonitorFetcher(
	schedulerRepo MonitorRepository,
	cfg config.MonitorChecksConfig,
	dispatcher Dispatcher,
) *MonitorFetcher {
	return &MonitorFetcher{
		baseScheduler: scheduler.New[Monitor](),
		repo:          schedulerRepo,
		dispatcher:    dispatcher,
		cfg:           cfg,
	}
}

func (s *MonitorFetcher) Run(ctx context.Context) error {
	if err := s.baseScheduler.Run(
		ctx,
		s.cfg.MonitorSchedulerTimeout,
		func(ctx context.Context) ([]Monitor, error) {
			return s.repo.ListDue(ctx, time.Now(), s.cfg.MonitorChecksLimit)
		},
		s.dispatcher.Submit,
	); err != nil {
		return err
	}

	return nil
}
