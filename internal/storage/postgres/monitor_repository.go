package postgres

import (
	"context"
	"errors"
	"url-monitor/internal/monitor"

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

	rows, err := r.pool.Query(ctx, query)
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
