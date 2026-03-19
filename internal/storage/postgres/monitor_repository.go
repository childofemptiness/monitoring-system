package postgres

import (
	"context"
	"url-monitor/internal/monitor"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewMonitorRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, monitor monitor.Monitor) (monitor.Monitor, error)
