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
		INSERT INTO monitors (url, interval_seconds)
		VALUES ($1, $2)
		RETURNING id, url, interval_seconds
	`

	var created monitor.Monitor
	err := r.pool.QueryRow(ctx, query, 
		m.URL, 
		m.IntervalSeconds,
	).Scan(
		&created.ID,
		&created.URL,
		&created.IntervalSeconds,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return monitor.Monitor{}, monitor.ErrMonitorAlreadyExists
		}

		return monitor.Monitor{}, err
	}

	return created, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return false
}
