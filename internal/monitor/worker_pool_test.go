package monitor

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
	"url-monitor/internal/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	workersCount = 1
	queueSize    = 3
)

type fakeProcessor struct {
	gotCtx            context.Context
	savedErr          error
	processedMonitors map[int64]Monitor
	mu                sync.Mutex
	wg                *sync.WaitGroup
}

func newFakeProcessor(monitorsCount int) *fakeProcessor {
	var wg sync.WaitGroup
	wg.Add(monitorsCount)

	return &fakeProcessor{
		processedMonitors: make(map[int64]Monitor, monitorsCount),
		wg:                &wg,
	}
}

func (fp *fakeProcessor) Process(ctx context.Context, monitor Monitor) error {
	fp.gotCtx = ctx

	fp.mu.Lock()
	fp.processedMonitors[monitor.ID] = monitor
	fp.mu.Unlock()

	fp.wg.Done()

	return fp.savedErr
}

func TestWorkerPool_Submit_Success(t *testing.T) {
	processor := &fakeProcessor{}
	reg := prometheus.NewRegistry()
	wp := NewWorkerPool(processor, workersCount, queueSize, metrics.NewMetrics(reg))

	monitor := newTestMonitor()
	ctx := context.Background()

	if err := wp.Submit(ctx, monitor); err != nil {
		t.Fatalf("failed to submit monitor %+v: %s", monitor, err)
	}

	received := <-wp.jobsCh

	if !reflect.DeepEqual(received, monitor) {
		t.Fatalf("failed to submit monitor: %+v", monitor)
	}
}

func TestWorkerPool_Submit_ContextCanceledError(t *testing.T) {
	processor := &fakeProcessor{}
	reg := prometheus.NewRegistry()
	wp := NewWorkerPool(processor, workersCount, queueSize, metrics.NewMetrics(reg))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	monitor := newTestMonitor()
	if err := wp.Submit(ctx, monitor); !errors.Is(err, context.Canceled) {
		t.Fatalf("failed to cancel monitor: %+v: %s", monitor, err)
	}
}

func TestWorkerPool_Submit_ContextDeadlineExceededError(t *testing.T) {
	processor := &fakeProcessor{}
	reg := prometheus.NewRegistry()
	wp := NewWorkerPool(processor, workersCount, queueSize, metrics.NewMetrics(reg))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	monitor := newTestMonitor()

	for i := 0; i < queueSize; i++ {
		if err := wp.Submit(ctx, monitor); err != nil {
			t.Fatalf("failed to submit monitor %+v: %s", monitor, err)
		}
	}

	if err := wp.Submit(ctx, monitor); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("failed to cancel monitor: %+v: %s", monitor, err)
	}
}

func TestWorkerPool_Run_Success(t *testing.T) {
	monitors := newTestMonitors()

	processor := newFakeProcessor(len(monitors))
	reg := prometheus.NewRegistry()
	wp := NewWorkerPool(processor, workersCount, queueSize, metrics.NewMetrics(reg))

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)

	go func(ctx context.Context) {
		errCh <- wp.Run(ctx)
	}(ctx)

	for _, monitor := range monitors {
		if err := wp.Submit(ctx, monitor); err != nil {
			t.Fatalf("failed to submit monitor %+v: %s", monitor, err)
		}
	}

	doneCh := make(chan struct{})
	go func() {
		processor.wg.Wait()
		close(doneCh)
	}()

	select {
	case <-doneCh:
	case <-time.After(2 * time.Second):
		t.Fatalf("failed to wait for worker pool to finish")
	}

	cancel()

	if err := <-errCh; err != nil {
		t.Fatalf("failed to cancel worker pool: %s", err)
	}

	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Fatalf("worker pool context error: got %v, want %v", ctx.Err(), context.Canceled)
	}

	for _, monitor := range monitors {
		if processed, ok := processor.processedMonitors[monitor.ID]; !ok || !reflect.DeepEqual(processed, monitor) {
			t.Fatalf("failed to processed monitor: %+v", monitor)
		}
	}
}

func TestWorkerPool_Run_Timeout(t *testing.T) {
	processor := newFakeProcessor(3)
	reg := prometheus.NewRegistry()
	wp := NewWorkerPool(processor, workersCount, queueSize, metrics.NewMetrics(reg))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	errCh := make(chan error, 1)

	go func(ctx context.Context) {
		errCh <- wp.Run(ctx)
	}(ctx)

	if err := <-errCh; err != nil {
		t.Fatalf("failed to cancel worker pool: %s", err)
	}

	if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("worker pool context error: got %v, want %v", ctx.Err(), context.DeadlineExceeded)
	}
}
