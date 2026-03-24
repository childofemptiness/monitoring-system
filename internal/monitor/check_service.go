package monitor

import (
	"context"
	"time"
)

type MonitorCheckRepository interface {
	CompleteCheck(ctx context.Context, check MonitorCheck, nextCheckAt time.Time) error
}

type CheckService struct {
	repo MonitorCheckRepository
}

func (c *CheckService) SaveCheckResult(ctx context.Context, check MonitorCheck, nextCheckAt time.Time) error {
	return c.repo.CompleteCheck(ctx, check, nextCheckAt)
}
