package monitor

import (
	"context"
	"net/url"
	"strings"
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

func (s *Service) CreateMonitor(ctx context.Context, input CreateMonitorInput) (Monitor, error) {
	if !s.validateURL(input.URL) {
		return Monitor{}, ErrInvalidURL
	}

	if input.IntervalSeconds <= 0 {
		return Monitor{}, ErrInvalidInterval
	}

	monitor := Monitor{
		URL: input.URL,
		IntervalSeconds: input.IntervalSeconds,
	}

	created, err := s.repo.Create(ctx, monitor)
	if err != nil {
		return Monitor{}, err
	}

	return created, nil
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
