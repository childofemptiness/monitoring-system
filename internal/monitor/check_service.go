package monitor

import (
	"context"
	"time"
)

type CheckRepository interface {
	CompleteCheck(ctx context.Context, check MonitorCheck, nextCheckAt time.Time) error
}

type CheckService struct {
	repo CheckRepository
}

func NewCheckService(repo CheckRepository) *CheckService {
	return &CheckService{repo: repo}
}

func (c *CheckService) SaveCheckResult(ctx context.Context, check MonitorCheck, nextCheckAt time.Time) error {
	return c.repo.CompleteCheck(ctx, check, nextCheckAt)
}
