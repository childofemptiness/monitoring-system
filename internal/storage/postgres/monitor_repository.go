package postgres

import (
	"context"
	"errors"
	"time"
	"url-monitor/internal/monitor"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewMonitorRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, m monitor.Monitor) (monitor.Monitor, error) {
	query := `
		INSERT INTO monitors (url, interval_seconds, next_check_at)
		VALUES ($1, $2, $3)
		RETURNING id, url, interval_seconds, created_at, updated_at, last_check_at, next_check_at
	`

	var created monitor.Monitor
	err := r.pool.QueryRow(ctx, query,
		m.URL,
		m.IntervalSeconds,
		m.NextCheckAt,
	).Scan(
		&created.ID,
		&created.URL,
		&created.IntervalSeconds,
		&created.CreatedAt,
		&created.UpdatedAt,
		&created.LastCheckAt,
		&created.NextCheckAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return monitor.Monitor{}, monitor.ErrMonitorAlreadyExists
		}

		return monitor.Monitor{}, err
	}

	return created, nil
}

func (r *Repository) List(ctx context.Context) ([]monitor.Monitor, error) {
	query := `
		SELECT
		    id,
		    url,
		    interval_seconds,
		    created_at,
		    updated_at,
		    last_check_at,
		    next_check_at
		FROM monitors
		ORDER BY id ASC
	`

	return r.executeQuery(ctx, query)
}

func (r *Repository) ListDue(ctx context.Context, now time.Time, limit int) ([]monitor.Monitor, error) {
	query := `
		SELECT 
			id, 
			url, 
			interval_seconds, 
			created_at, updated_at, 
			last_check_at, 
			next_check_at
		FROM monitors
		WHERE next_check_at <= $1
		ORDER BY id ASC
		LIMIT $2
	`

	return r.executeQuery(ctx, query, now, limit)
}

func (r *Repository) CompleteCheck(ctx context.Context, check monitor.MonitorCheck, nextCheckAt time.Time) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	query := `
		INSERT INTO monitor_checks (
		                            monitor_id, 
		                            status, 
		                            http_status_code, 
		                            error_message, 
		                            response_time_ms, 
		                            started_at, 
		                            finished_at
		                            )
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING 
			id, 
			monitor_id, 
			status, 
			http_status_code, 
			error_message, 
			response_time_ms, 
			started_at, 
			finished_at
	`

	_, err = tx.Exec(ctx, query,
		check.MonitorID,
		check.Status,
		check.HTTPStatusCode,
		check.ErrorMessage,
		check.ResponseTimeMS,
		check.StartedAt,
		check.FinishedAt,
	)
	if err != nil {
		return err
	}

	query = `
		UPDATE monitors
		SET last_check_at = $2,
		    next_check_at = $3
		WHERE id = $1
	`

	_, err = tx.Exec(ctx, query,
		check.MonitorID,
		check.FinishedAt,
		nextCheckAt,
	)
	if err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (r *Repository) executeQuery(ctx context.Context, query string, args ...any) ([]monitor.Monitor, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	monitors := make([]monitor.Monitor, 0)
	for rows.Next() {
		var m monitor.Monitor
		if err := rows.Scan(
			&m.ID,
			&m.URL,
			&m.IntervalSeconds,
			&m.CreatedAt,
			&m.UpdatedAt,
			&m.LastCheckAt,
			&m.NextCheckAt,
		); err != nil {
			return nil, err
		}

		monitors = append(monitors, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return monitors, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return false
}
