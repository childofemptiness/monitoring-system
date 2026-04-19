package check

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"
	"url-monitor/internal/monitor"
)

type fakeCheckRepository struct {
	savedCheck     monitor.MonitorCheck
	savedErr       error
	gotCtx         context.Context
	gotNextCheckAt time.Time
}

func (f *fakeCheckRepository) CompleteCheck(ctx context.Context, check monitor.MonitorCheck, nextCheckAt time.Time) error {
	f.savedCheck = check
	f.gotNextCheckAt = nextCheckAt
	f.gotCtx = ctx

	return f.savedErr
}

func TestCheckService_SaveCheckResult_Success(t *testing.T) {
	repo := &fakeCheckRepository{}
	svc := NewCheckStoreService(repo)

	ctx := context.Background()
	check := newTestCheck()
	nextCheckAt := time.Date(2026, time.March, 26, 12, 0, 45, 0, time.UTC)

	err := svc.SaveCheckResult(ctx, check, nextCheckAt)
	if err != nil {
		t.Fatalf("save check result: %v", err)
	}

	if repo.gotCtx != ctx {
		t.Errorf("save check result: got %v, want %v as ctx", repo.gotCtx, ctx)
	}

	if !repo.gotNextCheckAt.Equal(nextCheckAt) {
		t.Errorf("save check result: got %v, want %v as nextCheckAt", repo.gotNextCheckAt, nextCheckAt)
	}

	if !reflect.DeepEqual(repo.savedCheck, check) {
		t.Errorf("save check result: got %v, want %v", repo.savedCheck, check)
	}
}

func TestCheckService_SaveCheckResult_PropagateRepositoryError(t *testing.T) {
	repo := &fakeCheckRepository{}
	svc := NewCheckStoreService(repo)

	repo.savedErr = monitor.ErrMonitorNotFound

	ctx := context.Background()
	check := newTestCheck()
	nextCheckAt := time.Date(2026, time.March, 26, 12, 0, 45, 0, time.UTC)

	err := svc.SaveCheckResult(ctx, check, nextCheckAt)
	if !errors.Is(err, monitor.ErrMonitorNotFound) {
		t.Errorf("save check result: got %v, want ErrMonitorNotFound", err)
	}
}

func newTestCheck() monitor.MonitorCheck {
	finishedAt := time.Date(2026, time.March, 26, 12, 0, 0, 123456000, time.UTC)
	responseTimeMS := int64(100)
	startedAt := finishedAt.Add(-time.Duration(responseTimeMS) * time.Millisecond)

	return monitor.MonitorCheck{
		MonitorID:      1,
		Status:         monitor.MonitorCheckStatusUp,
		HTTPStatusCode: http.StatusOK,
		ErrorMessage:   "",
		ResponseTimeMS: responseTimeMS,
		StartedAt:      startedAt,
		FinishedAt:     finishedAt,
	}
}
