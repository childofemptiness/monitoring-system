package check

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-monitor/internal/monitor"
)

func TestCheckerChecker_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	checker := CheckRunner{}

	got := checker.Check(context.Background(), monitor.Monitor{
		URL:             srv.URL,
		IntervalSeconds: 10,
	})

	if got.Status != monitor.MonitorCheckStatusUp {
		t.Fatalf("expected status %q, got %q", monitor.MonitorCheckStatusUp, got.Status)
	}

	if got.HTTPStatusCode != http.StatusOK {
		t.Fatalf("expected http status %q, got %q", http.StatusOK, got.HTTPStatusCode)
	}
}

func TestCheckerChecker_Down(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	checker := CheckRunner{}

	got := checker.Check(context.Background(), monitor.Monitor{
		URL:             srv.URL,
		IntervalSeconds: 10,
	})

	if got.Status != monitor.MonitorCheckStatusDown {
		t.Fatalf("expected status %q, got %q", monitor.MonitorCheckStatusDown, got.Status)
	}

	if got.HTTPStatusCode != http.StatusInternalServerError {
		t.Fatalf("expected http status %q, got %q", http.StatusInternalServerError, got.HTTPStatusCode)
	}
}

func TestCheckerChecker_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	srv.Close()

	checker := CheckRunner{}

	got := checker.Check(context.Background(), monitor.Monitor{
		URL:             srv.URL,
		IntervalSeconds: 10,
	})

	if got.Status != monitor.MonitorCheckStatusError {
		t.Fatalf("expected status %q, got %q", monitor.MonitorCheckStatusError, got.Status)
	}

	if got.ErrorMessage == "" {
		t.Fatal("expected non-empty error message")
	}
}
