package monitor

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, monitor Monitor) (Monitor, error)
}

type Service struct {
	repo Repository
}

type CreateMonitorInput struct {
	URL     		string
	IntervalSeconds int
}

func NewMonitorService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateMonitor(ctx context.Context, input CreateMonitorInput) (Monitor, error)
