package check

import (
	"context"
	"time"
	"url-monitor/internal/metrics"
	"url-monitor/internal/monitor"
)

type Checker interface {
	Check(ctx context.Context, monitor monitor.Monitor) monitor.MonitorCheck
}
type CheckService interface {
	SaveCheckResult(ctx context.Context, check monitor.MonitorCheck, nextCheckAt time.Time) error
}
type CheckProcessor struct {
	checker      Checker
	checkService CheckService
	metrics      *metrics.Metrics
}

func NewCheckProcessor(checker Checker, checkService CheckService, metrics *metrics.Metrics) *CheckProcessor {
	return &CheckProcessor{
		checker:      checker,
		checkService: checkService,
		metrics:      metrics,
	}
}

func (cp *CheckProcessor) Process(ctx context.Context, monitor monitor.Monitor) error {
	cp.metrics.IncInflight()
	defer cp.metrics.DecInflight()

	check := cp.checker.Check(ctx, monitor)
	nextCheckAt := check.FinishedAt.Add(time.Duration(monitor.IntervalSeconds) * time.Second)

	if check.ErrorKind != "" {
		cp.metrics.IncRequestErrors(string(check.ErrorKind))
	}

	cp.metrics.ObserveCheck(string(check.Status), time.Duration(check.ResponseTimeMS)*time.Millisecond)

	err := cp.checkService.SaveCheckResult(ctx, check, nextCheckAt)
	if err != nil {
		return err
	}

	return nil
}
