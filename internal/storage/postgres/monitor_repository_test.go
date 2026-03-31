package postgres

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
	"url-monitor/internal/monitor"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRepositoryCreate(t *testing.T) {
	pool := setupTestDatabase(t)
	repo := NewMonitorRepository(pool)

	nextCheckAt := time.Now().UTC().Add(45 * time.Second).Truncate(time.Microsecond)

	created, err := repo.Create(context.Background(), monitor.Monitor{
		URL:             "https://example.com",
		IntervalSeconds: 45,
		NextCheckAt:     &nextCheckAt,
	})
	if err != nil {
		t.Fatalf("create monitor: %v", err)
	}

	if created.ID == 0 {
		t.Errorf("expected created monitor id to be set")
	}
	if created.CreatedAt.IsZero() {
		t.Errorf("expected created_at to be set")
	}
	if created.NextCheckAt == nil {
		t.Errorf("expected next_check_at to be set")
	}
	if !created.NextCheckAt.Equal(nextCheckAt) {
		t.Errorf("expected next_check_at %s, got %s", nextCheckAt, created.NextCheckAt)
	}
}

func TestRepositoryList(t *testing.T) {
	pool := setupTestDatabase(t)
	repo := NewMonitorRepository(pool)

	firstNextCheckAt := time.Now().UTC().Add(10 * time.Second).Truncate(time.Microsecond)
	secondNextCheckAt := time.Now().UTC().Add(20 * time.Second).Truncate(time.Microsecond)

	insertMonitorRow(t, pool, "https://first.example.com", 10, firstNextCheckAt)
	insertMonitorRow(t, pool, "https://second.example.com", 20, secondNextCheckAt)

	monitors, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("list monitors: %v", err)
	}

	if len(monitors) != 2 {
		t.Fatalf("expected 2 monitors, got %d", len(monitors))
	}
	if monitors[0].URL != "https://first.example.com" {
		t.Errorf("expected first monitor url to be https://first.example.com, got %s", monitors[0].URL)
	}
	if monitors[0].CreatedAt.IsZero() {
		t.Errorf("expected created_at to be populated")
	}
	if monitors[1].NextCheckAt == nil || !monitors[1].NextCheckAt.Equal(secondNextCheckAt) {
		t.Errorf("expected second next_check_at %s, got %v", secondNextCheckAt, monitors[1].NextCheckAt)
	}
}

func TestRepositoryListDue(t *testing.T) {
	pool := setupTestDatabase(t)
	repo := NewMonitorRepository(pool)

	now := time.Now().UTC().Truncate(time.Microsecond)

	firstDueID := insertMonitorRow(t, pool, "https://due-one.example.com", 10, now.Add(-2*time.Minute))
	secondDueID := insertMonitorRow(t, pool, "https://due-two.example.com", 20, now)
	insertMonitorRow(t, pool, "https://future.example.com", 30, now.Add(2*time.Minute))
	insertMonitorWithoutNextCheckAt(t, pool, "https://missing-next-check.example.com", 40)

	monitors, err := repo.ListDue(context.Background(), now, 10)
	if err != nil {
		t.Fatalf("list due monitors: %v", err)
	}

	if len(monitors) != 2 {
		t.Fatalf("expected 2 due monitors, got %d", len(monitors))
	}
	if monitors[0].ID != firstDueID {
		t.Errorf("expected first due monitor id %d, got %d", firstDueID, monitors[0].ID)
	}
	if monitors[1].ID != secondDueID {
		t.Errorf("expected second due monitor id %d, got %d", secondDueID, monitors[1].ID)
	}
}

func TestRepositoryListDueRespectsLimit(t *testing.T) {
	pool := setupTestDatabase(t)
	repo := NewMonitorRepository(pool)

	now := time.Now().UTC().Truncate(time.Microsecond)
	firstDueID := insertMonitorRow(t, pool, "https://due-limit-one.example.com", 10, now.Add(-time.Minute))
	insertMonitorRow(t, pool, "https://due-limit-two.example.com", 20, now.Add(-30*time.Second))

	monitors, err := repo.ListDue(context.Background(), now, 1)
	if err != nil {
		t.Fatalf("list due monitors with limit: %v", err)
	}

	if len(monitors) != 1 {
		t.Fatalf("expected 1 due monitor, got %d", len(monitors))
	}
	if monitors[0].ID != firstDueID {
		t.Errorf("expected monitor id %d, got %d", firstDueID, monitors[0].ID)
	}
}

func TestRepositoryCreateReturnsDuplicateError(t *testing.T) {
	pool := setupTestDatabase(t)
	repo := NewMonitorRepository(pool)

	createUniqueConstraint(t, pool)

	nextCheckAt := time.Now().UTC().Add(time.Minute).Truncate(time.Microsecond)
	input := monitor.Monitor{
		URL:             "https://duplicate.example.com",
		IntervalSeconds: 60,
		NextCheckAt:     &nextCheckAt,
	}

	if _, err := repo.Create(context.Background(), input); err != nil {
		t.Fatalf("create first monitor: %v", err)
	}

	_, err := repo.Create(context.Background(), input)
	if !errors.Is(err, monitor.ErrMonitorAlreadyExists) {
		t.Fatalf("expected ErrMonitorAlreadyExists, got %v", err)
	}
}

func TestRepository_CompleteCheckSuccessful(t *testing.T) {
	pool := setupTestDatabase(t)
	repo := NewMonitorRepository(pool)

	nextCheckAt := time.Now().UTC().Add(45 * time.Second).Truncate(time.Microsecond)

	createdMonitor, err := repo.Create(context.Background(), monitor.Monitor{
		URL:             "https://example.com",
		IntervalSeconds: 45,
		NextCheckAt:     &nextCheckAt,
	})

	if err != nil {
		t.Fatalf("create monitor: %v", err)
	}

	finishedAt := time.Now().UTC().Truncate(time.Microsecond)
	responseTimeMS := int64(100)

	duration := time.Duration(responseTimeMS) * time.Millisecond
	startedAt := finishedAt.Add(-duration).Truncate(time.Microsecond)

	check := monitor.MonitorCheck{
		MonitorID:      createdMonitor.ID,
		Status:         monitor.MonitorCheckStatusUp,
		HTTPStatusCode: http.StatusOK,
		ErrorMessage:   "",
		ResponseTimeMS: responseTimeMS,
		StartedAt:      startedAt,
		FinishedAt:     finishedAt,
	}

	err = repo.CompleteCheck(context.Background(), check, nextCheckAt)
	if err != nil {
		t.Fatalf("complete check: %v", err)
	}

	query := `
		SELECT
			id,
			monitor_id,
			status,
  			http_status_code,
  			error_kind,
			error_message,
			response_time_ms,
			started_at,
			finished_at
		FROM monitor_checks
		WHERE
		    monitor_id 		 = $1 AND
			status 	         = $2 AND 
		    http_status_code = $3 AND 
		    error_kind = 	   $4 AND
		    error_message    = $5 AND
		    response_time_ms = $6 AND
		    started_at       = $7 AND 
		    finished_at      = $8
   `
	var createdCheck monitor.MonitorCheck
	err = pool.QueryRow(context.Background(), query,
		check.MonitorID,
		check.Status,
		check.HTTPStatusCode,
		check.ErrorKind,
		check.ErrorMessage,
		check.ResponseTimeMS,
		check.StartedAt,
		check.FinishedAt,
	).Scan(
		&createdCheck.ID,
		&createdCheck.MonitorID,
		&createdCheck.Status,
		&createdCheck.HTTPStatusCode,
		&createdCheck.ErrorKind,
		&createdCheck.ErrorMessage,
		&createdCheck.ResponseTimeMS,
		&createdCheck.StartedAt,
		&createdCheck.FinishedAt,
	)
	if err != nil {
		t.Fatalf("get created check: %v", err)
	}

	if createdCheck.ID == 0 {
		t.Errorf("expected created monitor check id to be set")
	}

	if createdCheck.MonitorID != createdMonitor.ID {
		t.Errorf("expected created check monitor id %d, got %d", createdMonitor.ID, createdCheck.ID)
	}

	if createdCheck.Status != monitor.MonitorCheckStatusUp {
		t.Errorf("expected created monitor check status %s, got %s", monitor.MonitorCheckStatusUp, createdCheck.Status)
	}

	if createdCheck.ErrorMessage != "" {
		t.Errorf("expected created monitor check status to be empty, got %s", createdCheck.ErrorMessage)
	}

	if createdCheck.ResponseTimeMS != responseTimeMS {
		t.Errorf("expected created monitor check response time to be %d, got %d", responseTimeMS, createdCheck.ResponseTimeMS)
	}

	if !createdCheck.StartedAt.Equal(startedAt) {
		t.Errorf("expected created monitor check created at time to be %s, got %v", startedAt, createdCheck.StartedAt)
	}

	if !createdCheck.FinishedAt.Equal(finishedAt) {
		t.Errorf("expected created monitor check finished at time to be %s, got %v", finishedAt, createdCheck.FinishedAt)
	}

	m := selectMonitorById(t, pool, createdCheck.MonitorID)

	if !(*(m.LastCheckAt)).Equal(finishedAt) {
		t.Errorf("expected monitor last check at time to be %s, got %v", finishedAt, m.LastCheckAt)
	}

	if !(*(m.NextCheckAt)).Equal(nextCheckAt) {
		t.Errorf("expected monitor next check at time to be %s, got %v", nextCheckAt, m.NextCheckAt)
	}
}

func TestRepository_CompleteCheckNonExistentMonitorIDInsertError(t *testing.T) {
	pool := setupTestDatabase(t)
	repo := NewMonitorRepository(pool)

	nextCheckAt := time.Now().UTC().Add(45 * time.Second).Truncate(time.Microsecond)

	finishedAt := time.Now().UTC().Truncate(time.Microsecond)
	responseTimeMS := int64(100)

	duration := time.Duration(responseTimeMS) * time.Millisecond
	startedAt := finishedAt.Add(-duration).Truncate(time.Microsecond)

	check := monitor.MonitorCheck{
		MonitorID:      int64(111),
		Status:         monitor.MonitorCheckStatusUp,
		HTTPStatusCode: http.StatusOK,
		ErrorMessage:   "",
		ResponseTimeMS: responseTimeMS,
		StartedAt:      startedAt,
		FinishedAt:     finishedAt,
	}

	err := repo.CompleteCheck(context.Background(), check, nextCheckAt)
	if !errors.Is(err, monitor.ErrMonitorNotFound) {
		t.Errorf("expected ErrMonitorNotFound error, got %v", err)
	}
}

func insertMonitorRow(t *testing.T, pool *pgxpool.Pool, rawURL string, intervalSeconds int, nextCheckAt time.Time) int64 {
	t.Helper()

	var id int64
	err := pool.QueryRow(context.Background(), `
		INSERT INTO monitors (url, interval_seconds, next_check_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`, rawURL, intervalSeconds, nextCheckAt).Scan(&id)
	if err != nil {
		t.Fatalf("insert monitor row: %v", err)
	}

	return id
}

func insertMonitorWithoutNextCheckAt(t *testing.T, pool *pgxpool.Pool, rawURL string, intervalSeconds int) int64 {
	t.Helper()

	var id int64
	err := pool.QueryRow(context.Background(), `
		INSERT INTO monitors (url, interval_seconds)
		VALUES ($1, $2)
		RETURNING id
	`, rawURL, intervalSeconds).Scan(&id)
	if err != nil {
		t.Fatalf("insert monitor without next_check_at: %v", err)
	}

	return id
}

func createUniqueConstraint(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	if _, err := pool.Exec(context.Background(), `
		ALTER TABLE monitors
		ADD CONSTRAINT monitors_url_key UNIQUE (url)
	`); err != nil {
		t.Fatalf("create unique constraint: %v", err)
	}
}

func selectMonitorById(t *testing.T, pool *pgxpool.Pool, id int64) monitor.Monitor {
	t.Helper()

	var m monitor.Monitor
	err := pool.QueryRow(context.Background(), `
		SELECT 
		    id, 
		    url, 
		    interval_seconds,
		    last_check_at,
		    next_check_at
		FROM monitors
		where id = $1
    `, id).Scan(
		&m.ID,
		&m.URL,
		&m.IntervalSeconds,
		&m.LastCheckAt,
		&m.NextCheckAt,
	)
	if err != nil {
		t.Fatalf("select monitor by id: %v", err)
	}

	return m
}
