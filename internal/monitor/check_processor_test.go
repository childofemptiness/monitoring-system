package monitor

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"
	"url-monitor/internal/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

type fakeChecker struct {
	gotCtx     context.Context
	gotMonitor Monitor
}

type fakeCheckService struct {
	gotCtx          context.Context
	savedCheck      MonitorCheck
	savedNextTimeAt time.Time
	savedErr        error
}

func (fc *fakeChecker) Check(ctx context.Context, monitor Monitor) MonitorCheck {
	fc.gotCtx = ctx
	fc.gotMonitor = monitor

	_, check := newTestData()

	return check
}

func (fcs *fakeCheckService) SaveCheckResult(ctx context.Context, check MonitorCheck, nextCheckAt time.Time) error {
	fcs.gotCtx = ctx
	fcs.savedCheck = check
	fcs.savedNextTimeAt = nextCheckAt

	return fcs.savedErr
}

func TestCheckProcessor_Process_Success(t *testing.T) {
	checker := &fakeChecker{}
	checkService := &fakeCheckService{}
	reg := prometheus.NewRegistry()
	checkProcessor := NewCheckProcessor(checker, checkService, metrics.NewMetrics(reg))

	ctx := context.Background()
	monitor, _ := newTestData()

	if err := checkProcessor.Process(ctx, monitor); err != nil {
		t.Fatalf("checkProcessor.Process(): %v", err)
	}

	if checker.gotCtx != ctx {
		t.Errorf("checkProcessor.Process(): got context %v; want %v", checker.gotCtx, ctx)
	}

	if !reflect.DeepEqual(checker.gotMonitor, monitor) {
		t.Errorf("checkProcessor.Process(): got monitor %v; want %v", checker.gotMonitor, monitor)
	}

	if checkService.gotCtx != ctx {
		t.Errorf("checkProcessor.Process(): got context %v; want %v", checkService.gotCtx, ctx)
	}

	if checkService.savedCheck.MonitorID != monitor.ID {
		t.Errorf("checkProcessor.Process(): monitor ID should be %v; got %v", monitor.ID, checkService.savedCheck.MonitorID)
	}

	neededNextCheckAt := checkService.savedCheck.FinishedAt.Add(time.Duration(monitor.IntervalSeconds) * time.Second)

	if !checkService.savedNextTimeAt.Equal(neededNextCheckAt) {
		t.Errorf("checkProcess.Process(): nextCheckAt should be %v; got %v", monitor.NextCheckAt, checkService.savedNextTimeAt)
	}
}

func TestCheckProcessor_Process_SaveCheckResultError(t *testing.T) {
	checker := &fakeChecker{}
	checkService := &fakeCheckService{}
	reg := prometheus.NewRegistry()
	checkProcessor := NewCheckProcessor(checker, checkService, metrics.NewMetrics(reg))

	checkService.savedErr = ErrMonitorNotFound

	ctx := context.Background()

	monitor, _ := newTestData()

	if err := checkProcessor.Process(ctx, monitor); !errors.Is(err, ErrMonitorNotFound) {
		t.Fatalf("checkProcessor.Process(): got %v; want %v", err, ErrMonitorNotFound)
	}

	if checker.gotCtx != ctx {
		t.Errorf("checkProcessor.Process(): got context %v; want %v", checker.gotCtx, ctx)
	}

	if !reflect.DeepEqual(checker.gotMonitor, monitor) {
		t.Errorf("checkProcessor.Process(): got monitor %v; want %v", monitor, checker.gotMonitor)
	}

	if checkService.gotCtx != ctx {
		t.Errorf("checkProcessor.Process(): got %v; want %v", checkService.gotCtx, ctx)
	}
}

func newTestData() (Monitor, MonitorCheck) {
	responseTimeMS := int64(100)
	finishedAt := time.Date(2026, 3, 27, 15, 54, 11, 0, time.UTC)
	startedAt := finishedAt.Add(-time.Duration(responseTimeMS) * time.Millisecond)

	monitor := Monitor{
		ID:              1,
		URL:             "https://example1.com",
		IntervalSeconds: 10,
		CreatedAt:       time.Date(2026, 3, 27, 15, 50, 0, 0, time.UTC),
		NextCheckAt:     &startedAt,
	}

	check := MonitorCheck{
		MonitorID:      monitor.ID,
		Status:         MonitorCheckStatusUp,
		HTTPStatusCode: http.StatusOK,
		ErrorMessage:   "",
		ResponseTimeMS: responseTimeMS,
		StartedAt:      startedAt,
		FinishedAt:     finishedAt,
	}

	return monitor, check
}
