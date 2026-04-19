package monitor

import (
	"context"
	"net/url"
	"strings"
	"time"
)

type Repository interface {
	Create(ctx context.Context, monitor Monitor) (Monitor, error)
	List(ctx context.Context) ([]Monitor, error)
}

type Service struct {
	repo Repository
}

type CreateMonitorInput struct {
	URL             string
	IntervalSeconds int
}

func NewMonitorService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateMonitor(ctx context.Context, input CreateMonitorInput) (Monitor, error) {
	if !s.validateURL(input.URL) {
		return Monitor{}, ErrInvalidURL
	}

	if input.IntervalSeconds <= 0 {
		return Monitor{}, ErrInvalidInterval
	}

	nextCheckAt := time.Now().Add(time.Duration(input.IntervalSeconds) * time.Second)

	monitor := Monitor{
		URL:             input.URL,
		IntervalSeconds: input.IntervalSeconds,
		NextCheckAt:     &nextCheckAt,
	}

	created, err := s.repo.Create(ctx, monitor)
	if err != nil {
		return Monitor{}, err
	}

	return created, nil
}

func (s *Service) ListMonitors(ctx context.Context) ([]Monitor, error) {
	monitors, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	return monitors, nil
}

func (s *Service) validateURL(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}

	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return false
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	if u.Host == "" {
		return false
	}

	return true
}
