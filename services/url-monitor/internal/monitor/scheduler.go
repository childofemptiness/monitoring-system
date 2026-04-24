package monitor

import (
	"context"
	"log"
	"time"
)

const (
	checksLimit = 5
)

type SchedulerRepository interface {
	ListDue(ctx context.Context, now time.Time, limit int) ([]Monitor, error)
}

type Publisher interface {
	Submit(ctx context.Context, m Monitor) error
}
type Scheduler struct {
	repo      SchedulerRepository
	publisher Publisher
	timeout   time.Duration
}

func NewScheduler(
	schedulerRepo SchedulerRepository,
	publisher Publisher,
	timeout time.Duration,
) *Scheduler {
	return &Scheduler{
		repo:      schedulerRepo,
		publisher: publisher,
		timeout:   timeout,
	}
}

func (s *Scheduler) Run(ctx context.Context) error {
	log.Printf("scheduler started")

	ticker := time.NewTicker(s.timeout)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("scheduler stopped")
			return nil
		case <-ticker.C:
			if err := s.runOnce(ctx); err != nil {
				log.Printf("scheduler run error: %v", err)
				return err
			}
		}
	}
}

func (s *Scheduler) runOnce(ctx context.Context) error {
	monitors, err := s.repo.ListDue(ctx, time.Now(), checksLimit)
	if err != nil {
		return err
	}

	for _, monitor := range monitors {
		select {
		case <-ctx.Done():
			return nil
		default:
			err := s.publisher.Submit(ctx, monitor)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
