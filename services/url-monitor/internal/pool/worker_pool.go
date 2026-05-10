package pool

import (
	"context"
	"errors"
	"sync"
)

type Processor[T any] interface {
	Process(ctx context.Context, item T) error
}

type WorkerPool[T any] struct {
	processor    Processor[T]
	workersCount int
	jobsCh       chan T
}

func NewWorkerPool[T any](
	processor Processor[T],
	workersCount int,
	queueSize int,
) *WorkerPool[T] {

	if processor == nil {
		panic("nil processor")
	}

	if workersCount < 1 {
		panic("workers count must be greater than zero")
	}

	if queueSize < 1 {
		panic("queueSize must be greater than zero")
	}

	return &WorkerPool[T]{
		processor:    processor,
		workersCount: workersCount,
		jobsCh:       make(chan T, queueSize),
	}
}

func (wp *WorkerPool[T]) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	for i := 0; i < wp.workersCount; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()
			wp.runWorker(ctx, workerID)
		}(i)
	}

	<-ctx.Done()
	wg.Wait()

	return nil
}

func (wp *WorkerPool[T]) Submit(ctx context.Context, item T) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case wp.jobsCh <- item:
		return nil
	}
}

func (wp *WorkerPool[T]) runWorker(ctx context.Context, workerID int) {
	for {
		select {
		case <-ctx.Done():
			return

		case item, ok := <-wp.jobsCh:
			if !ok {
				return
			}

			if err := wp.processor.Process(ctx, item); err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
			}
		}
	}
}
