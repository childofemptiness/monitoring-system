package pool

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

type TestItem struct {
	ID int
}

type fakeProcessor struct {
	gotCtx         context.Context
	savedErr       error
	processedItems map[int]TestItem
	mu             sync.Mutex
	wg             *sync.WaitGroup
}

func newFakeProcessor(itemsCount int) *fakeProcessor {
	var wg sync.WaitGroup
	wg.Add(itemsCount)

	return &fakeProcessor{
		processedItems: make(map[int]TestItem, itemsCount),
		wg:             &wg,
	}
}

func (fp *fakeProcessor) Process(ctx context.Context, item TestItem) error {
	fp.gotCtx = ctx

	fp.mu.Lock()
	fp.processedItems[item.ID] = item
	fp.mu.Unlock()

	fp.wg.Done()

	return fp.savedErr
}

func TestWorkerPool_Submit_Success(t *testing.T) {
	processor := &fakeProcessor{}
	reg := prometheus.NewRegistry()
	wp := NewWorkerPool(processor, workersCount, queueSize, metrics.NewMetrics(reg))

	item := newTestItem()
	ctx := context.Background()

	if err := wp.Submit(ctx, item); err != nil {
		t.Fatalf("failed to submit monitor %+v: %s", item, err)
	}

	received := <-wp.jobsCh

	if !reflect.DeepEqual(received, item) {
		t.Fatalf("failed to submit item: %+v", item)
	}
}

func TestWorkerPool_Submit_ContextCanceledError(t *testing.T) {
	processor := &fakeProcessor{}
	reg := prometheus.NewRegistry()
	wp := NewWorkerPool(processor, workersCount, queueSize, metrics.NewMetrics(reg))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	item := newTestItem()
	if err := wp.Submit(ctx, item); !errors.Is(err, context.Canceled) {
		t.Fatalf("failed to cancel item: %+v: %s", item, err)
	}
}

func TestWorkerPool_Submit_ContextDeadlineExceededError(t *testing.T) {
	processor := &fakeProcessor{}
	reg := prometheus.NewRegistry()
	wp := NewWorkerPool(processor, workersCount, queueSize, metrics.NewMetrics(reg))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	item := newTestItem()

	for i := 0; i < queueSize; i++ {
		if err := wp.Submit(ctx, item); err != nil {
			t.Fatalf("failed to submit item %+v: %s", item, err)
		}
	}

	if err := wp.Submit(ctx, item); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("failed to cancel item: %+v: %s", item, err)
	}
}

func TestWorkerPool_Run_Success(t *testing.T) {
	items := newTestItems()

	processor := newFakeProcessor(len(items))
	reg := prometheus.NewRegistry()
	wp := NewWorkerPool(processor, workersCount, queueSize, metrics.NewMetrics(reg))

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)

	go func(ctx context.Context) {
		errCh <- wp.Run(ctx)
	}(ctx)

	for _, item := range items {
		if err := wp.Submit(ctx, item); err != nil {
			t.Fatalf("failed to submit item %+v: %s", item, err)
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

	for _, item := range items {
		if processed, ok := processor.processedItems[item.ID]; !ok || !reflect.DeepEqual(processed, item) {
			t.Fatalf("failed to processed item: %+v", item)
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

func newTestItem() TestItem {
	return TestItem{
		ID: 1,
	}
}

func newTestItems() []TestItem {
	return []TestItem{
		{ID: 1},
		{ID: 2},
		{ID: 3},
	}
}
