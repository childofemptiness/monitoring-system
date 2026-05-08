package scheduler

import (
	"context"
	"time"
)

type Scheduler[T any] struct{}

func New[T any]() Scheduler[T] {
	return Scheduler[T]{}
}

func (s *Scheduler[T]) Run(
	ctx context.Context,
	timeout time.Duration,
	fetcher func(context.Context) ([]T, error),
	dispatcher func(context.Context, T) error,
) error {
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := s.runOnce(ctx, fetcher, dispatcher); err != nil {
				return err
			}
		}
	}
}

func (s *Scheduler[T]) runOnce(
	ctx context.Context,
	fetcher func(ctx context.Context) ([]T, error),
	dispatcher func(ctx context.Context, item T) error,
) error {
	fetchedItems, err := fetcher(ctx)
	if err != nil {
		return err
	}

	for _, item := range fetchedItems {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := dispatcher(ctx, item); err != nil {
				return err
			}
		}
	}

	return nil
}
