package monitor

import (
	"context"
	"errors"
	"testing"
	"time"
)

const timeInterval = 2 * time.Millisecond

type fakeSchedulerRepository struct {
	gotCtx   context.Context
	called   chan struct{}
	gotTime  time.Time
	gotLimit int
	savedErr error
}

type fakePublisher struct {
	gotCtx            context.Context
	submittedMonitors map[int64]Monitor
	savedErr          error
}

func newFakeSchedulerRepository() *fakeSchedulerRepository {
	return &fakeSchedulerRepository{
		called: make(chan struct{}, 1),
	}
}

func newFakePublisher(monitorsCount int) *fakePublisher {
	return &fakePublisher{
		submittedMonitors: make(map[int64]Monitor, monitorsCount),
	}
}

func (fsr *fakeSchedulerRepository) ListDue(ctx context.Context, now time.Time, limit int) ([]Monitor, error) {
	fsr.gotCtx = ctx
	fsr.gotTime = now
	fsr.gotLimit = limit

	select {
	case fsr.called <- struct{}{}:
	default:
	}

	return newTestMonitors(), fsr.savedErr
}

func (fp *fakePublisher) Submit(ctx context.Context, monitor Monitor) error {
	fp.gotCtx = ctx

	fp.submittedMonitors[monitor.ID] = monitor

	return fp.savedErr
}

func TestScheduler_Run_SucceedContextCancellation(t *testing.T) {
	monitors := newTestMonitors()

	repo := newFakeSchedulerRepository()
	publisher := newFakePublisher(len(monitors))

	scheduler := NewScheduler(repo, publisher, timeInterval)
	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)

	go func(ctx context.Context) {
		errCh <- scheduler.Run(ctx)
	}(ctx)

	select {
	case <-repo.called:
		cancel()
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for scheduler to finish")
	}

	if err := <-errCh; err != nil {
		t.Fatalf("scheduler run failed: %v", err)
	}

	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Fatalf("scheduler context error: got %v, want %v", ctx.Err(), context.Canceled)
	}

	if repo.gotCtx != ctx {
		t.Errorf("scheduler run failed: got %v, want %v", repo.gotCtx, ctx)
	}

	if repo.gotTime.IsZero() {
		t.Errorf("scheduler run failed: gotTime should be set")
	}

	if repo.gotLimit != checksLimit {
		t.Errorf("scheduler run failed: gotLimit %v, want %v", repo.gotLimit, checksLimit)
	}
}

func TestScheduler_Run_ListDueReturnedError(t *testing.T) {
	repo := newFakeSchedulerRepository()
	publisher := newFakePublisher(3)
	scheduler := NewScheduler(repo, publisher, timeInterval)

	repoErr := errors.New("repo failed")
	repo.savedErr = repoErr

	ctx := context.Background()

	errCh := make(chan error, 1)
	go func(ctx context.Context) {
		errCh <- scheduler.Run(ctx)
	}(ctx)

	select {
	case err := <-errCh:
		if !errors.Is(err, repoErr) {
			t.Fatalf("scheduler run failed: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for scheduler to finish")
	}
}

func TestScheduler_Run_SubmitReturnedError(t *testing.T) {
	repo := newFakeSchedulerRepository()
	publisher := newFakePublisher(3)
	scheduler := NewScheduler(repo, publisher, timeInterval)

	publisherErr := errors.New("submit failed")
	publisher.savedErr = publisherErr

	ctx := context.Background()
	errCh := make(chan error, 1)
	go func(ctx context.Context) {
		errCh <- scheduler.Run(ctx)
	}(ctx)

	select {
	case err := <-errCh:
		if !errors.Is(err, publisherErr) {
			t.Fatalf("scheduler run failed: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for scheduler to finish")
	}
}

func TestScheduler_Run_TimeoutError(t *testing.T) {
	monitors := newTestMonitors()

	repo := newFakeSchedulerRepository()
	publisher := newFakePublisher(len(monitors))
	scheduler := NewScheduler(repo, publisher, timeInterval)

	ctx, cancel := context.WithTimeout(context.Background(), timeInterval)
	defer cancel()

	errCh := make(chan error, 1)
	go func(ctx context.Context) {
		errCh <- scheduler.Run(ctx)
	}(ctx)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("scheduler run failed: %v", err)
		}

		if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
			t.Fatalf("scheduler context error: got %v, want %v", ctx.Err(), context.DeadlineExceeded)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for scheduler to finish")
	}
}
