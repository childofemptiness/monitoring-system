package check

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"
	"url-monitor/internal/metrics"
	"url-monitor/internal/monitor"

	"github.com/prometheus/client_golang/prometheus"
)

type fakeChecker struct {
	gotCtx     context.Context
	gotMonitor monitor.Monitor
}

type fakeCheckService struct {
	gotCtx          context.Context
	savedCheck      monitor.MonitorCheck
	savedNextTimeAt time.Time
	savedErr        error
}

func (fc *fakeChecker) Check(ctx context.Context, monitor monitor.Monitor) monitor.MonitorCheck {
	fc.gotCtx = ctx
	fc.gotMonitor = monitor

	_, check := newTestData()

	return check
}

func (fcs *fakeCheckService) SaveCheckResult(ctx context.Context, check monitor.MonitorCheck, nextCheckAt time.Time) error {
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
	m, _ := newTestData()

	if err := checkProcessor.Process(ctx, m); err != nil {
		t.Fatalf("checkProcessor.Process(): %v", err)
	}

	if checker.gotCtx != ctx {
		t.Errorf("checkProcessor.Process(): got context %v; want %v", checker.gotCtx, ctx)
	}

	if !reflect.DeepEqual(checker.gotMonitor, m) {
		t.Errorf("checkProcessor.Process(): got monitor %v; want %v", checker.gotMonitor, m)
	}

	if checkService.gotCtx != ctx {
		t.Errorf("checkProcessor.Process(): got context %v; want %v", checkService.gotCtx, ctx)
	}

	if checkService.savedCheck.MonitorID != m.ID {
		t.Errorf("checkProcessor.Process(): monitor ID should be %v; got %v", m.ID, checkService.savedCheck.MonitorID)
	}

	neededNextCheckAt := checkService.savedCheck.FinishedAt.Add(time.Duration(m.IntervalSeconds) * time.Second)

	if !checkService.savedNextTimeAt.Equal(neededNextCheckAt) {
		t.Errorf("checkProcess.Process(): nextCheckAt should be %v; got %v", m.NextCheckAt, checkService.savedNextTimeAt)
	}
}

func TestCheckProcessor_Process_SaveCheckResultError(t *testing.T) {
	checker := &fakeChecker{}
	checkService := &fakeCheckService{}
	reg := prometheus.NewRegistry()
	checkProcessor := NewCheckProcessor(checker, checkService, metrics.NewMetrics(reg))

	checkService.savedErr = monitor.ErrMonitorNotFound

	ctx := context.Background()

	m, _ := newTestData()

	if err := checkProcessor.Process(ctx, m); !errors.Is(err, monitor.ErrMonitorNotFound) {
		t.Fatalf("checkProcessor.Process(): got %v; want %v", err, monitor.ErrMonitorNotFound)
	}

	if checker.gotCtx != ctx {
		t.Errorf("checkProcessor.Process(): got context %v; want %v", checker.gotCtx, ctx)
	}

	if !reflect.DeepEqual(checker.gotMonitor, m) {
		t.Errorf("checkProcessor.Process(): got monitor %v; want %v", m, checker.gotMonitor)
	}

	if checkService.gotCtx != ctx {
		t.Errorf("checkProcessor.Process(): got %v; want %v", checkService.gotCtx, ctx)
	}
}

func newTestData() (monitor.Monitor, monitor.MonitorCheck) {
	responseTimeMS := int64(100)
	finishedAt := time.Date(2026, 3, 27, 15, 54, 11, 0, time.UTC)
	startedAt := finishedAt.Add(-time.Duration(responseTimeMS) * time.Millisecond)

	m := monitor.Monitor{
		ID:              1,
		URL:             "https://example1.com",
		IntervalSeconds: 10,
		CreatedAt:       time.Date(2026, 3, 27, 15, 50, 0, 0, time.UTC),
		NextCheckAt:     &startedAt,
	}

	check := monitor.MonitorCheck{
		MonitorID:      m.ID,
		Status:         monitor.MonitorCheckStatusUp,
		HTTPStatusCode: http.StatusOK,
		ErrorMessage:   "",
		ResponseTimeMS: responseTimeMS,
		StartedAt:      startedAt,
		FinishedAt:     finishedAt,
	}

	return m, check
}
