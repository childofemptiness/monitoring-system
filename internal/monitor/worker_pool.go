package monitor

import (
	"context"
	"errors"
	"log"
	"sync"
)

type Processor interface {
	Process(ctx context.Context, monitor Monitor) error
}
type WorkerPool struct {
	processor    Processor
	workersCount int
	jobsCh       chan Monitor
}

func NewWorkerPool(
	processor Processor,
	workersCount int,
	queueSize int,
) *WorkerPool {

	if processor == nil {
		panic("nil processor")
	}

	if workersCount < 1 {
		panic("workers count must be greater than zero")
	}

	if queueSize < 1 {
		panic("queueSize must be greater than zero")
	}

	return &WorkerPool{
		processor:    processor,
		workersCount: workersCount,
		jobsCh:       make(chan Monitor, queueSize),
	}
}

func (wp *WorkerPool) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	for i := 0; i < wp.workersCount; i++ {
		wg.Add(1)

		go func(ctx context.Context, wg *sync.WaitGroup, workerID int) {
			defer wg.Done()
			wp.runWorker(ctx, workerID)
		}(ctx, &wg, i)
	}

	<-ctx.Done()
	wg.Wait()

	return ctx.Err()
}

func (wp *WorkerPool) Submit(ctx context.Context, monitor Monitor) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case wp.jobsCh <- monitor:
		return nil
	}
}

func (wp *WorkerPool) runWorker(ctx context.Context, workerID int) {
	log.Printf("worker %d starting", workerID)

	for {
		select {
		case <-ctx.Done():
			return

		case monitor, ok := <-wp.jobsCh:
			if !ok {
				return
			}

			if err := wp.processor.Process(ctx, monitor); err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
			}
		}
	}
}
