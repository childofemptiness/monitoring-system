package monitor

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type fakeRepository struct {
	createdMonitor  Monitor
	createdError    error
	createCalled    bool
	returnEmptyList bool
	gotCtx          context.Context
}

func (f *fakeRepository) Create(ctx context.Context, m Monitor) (Monitor, error) {
	if f.createdMonitor.URL == m.URL && f.createdMonitor.IntervalSeconds == m.IntervalSeconds {
		return f.createdMonitor, ErrMonitorAlreadyExists
	}

	f.createdMonitor = m
	f.gotCtx = ctx
	f.createCalled = true
	return f.createdMonitor, f.createdError
}

func (f *fakeRepository) List(ctx context.Context) ([]Monitor, error) {
	f.gotCtx = ctx

	if f.returnEmptyList {
		return []Monitor{}, nil
	}

	return newTestMonitors(), nil
}

func TestService_CreateMonitor_Success(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewMonitorService(repo)

	input := CreateMonitorInput{
		URL:             "https://example.com",
		IntervalSeconds: 45,
	}

	ctx := context.Background()

	created, err := svc.CreateMonitor(ctx, input)
	if err != nil {
		t.Errorf("CreateMonitor() error = %v, want nil", err)
	}

	if !repo.createCalled {
		t.Errorf("Repository.Create() was not called")
	}

	if repo.gotCtx != ctx {
		t.Errorf("CreateMonitor() got = %v, want %v", repo.gotCtx, ctx)
	}

	if repo.createdMonitor.URL != input.URL {
		t.Errorf("CreateMonitor() got = %v, want %v", repo.createdMonitor.URL, input.URL)
	}

	if repo.createdMonitor.IntervalSeconds != input.IntervalSeconds {
		t.Errorf("CreateMonitor() got = %v, want %v", repo.createdMonitor.IntervalSeconds, input.IntervalSeconds)
	}

	if created.URL != input.URL {
		t.Errorf("CreateMonitor() URL = %v, want %v", created.URL, input.URL)
	}

	if created.IntervalSeconds != input.IntervalSeconds {
		t.Errorf("CreateMonitor() IntervalSeconds = %v, want %v", created.IntervalSeconds, input.IntervalSeconds)
	}
}

func TestService_CreateMonitor_InvalidURLError(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewMonitorService(repo)

	input := CreateMonitorInput{
		URL:             "htt://example.com",
		IntervalSeconds: 45,
	}

	ctx := context.Background()

	_, err := svc.CreateMonitor(ctx, input)
	if !errors.Is(err, ErrInvalidURL) {
		t.Errorf("CreateMonitor() error = %v, want monitor.ErrInvalidURL", err)
	}
}

func TestService_CreateMonitor_InvalidIntervalError(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewMonitorService(repo)

	input := CreateMonitorInput{
		URL:             "https://example.com",
		IntervalSeconds: 0,
	}

	ctx := context.Background()

	_, err := svc.CreateMonitor(ctx, input)
	if !errors.Is(err, ErrInvalidInterval) {
		t.Errorf("CreateMonitor() error = %v, want monitor.ErrInvalidInterval", err)
	}
}

func TestService_CreateMonitor_DuplicateMonitorError(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewMonitorService(repo)

	input := CreateMonitorInput{
		URL:             "https://example.com",
		IntervalSeconds: 45,
	}

	ctx := context.Background()
	_, err := svc.CreateMonitor(ctx, input)
	if err != nil {
		t.Fatalf("CreateMonitor() error = %v, want nil", err)
	}

	_, err = svc.CreateMonitor(ctx, input)

	if !errors.Is(err, ErrMonitorAlreadyExists) {
		t.Errorf("CreateMonitor() error = %v, want ErrMonitorAlreadyExists", err)
	}
}

func TestService_ListMonitors_Success(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewMonitorService(repo)
	monitors := newTestMonitors()

	ctx := context.Background()
	receivedMonitors, err := svc.ListMonitors(ctx)
	if err != nil {
		t.Fatalf("ListMonitors() error = %v, want nil", err)
	}

	if repo.gotCtx != ctx {
		t.Errorf("ListMonitors() got = %v, want %v", repo.gotCtx, ctx)
	}

	if len(receivedMonitors) != len(monitors) {
		t.Errorf("ListMonitors() got = %v, want %v monitors", len(receivedMonitors), len(monitors))
	}

	if !reflect.DeepEqual(receivedMonitors, monitors) {
		t.Errorf("ListMonitors() got = %v, want %v", receivedMonitors, monitors)
	}
}

func TestService_ListMonitorsEmptyResult(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewMonitorService(repo)

	repo.returnEmptyList = true

	ctx := context.Background()
	receivedMonitors, err := svc.ListMonitors(ctx)
	if err != nil {
		t.Errorf("ListMonitors() error = %v, want nil", err)
	}

	if len(receivedMonitors) != 0 {
		t.Errorf("ListMonitors() got = %v, want %v monitors", len(receivedMonitors), 0)
	}
}
