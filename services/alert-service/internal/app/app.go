package app

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/childofemptiness/alert-service/internal/config"
	"github.com/childofemptiness/alert-service/internal/storage/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	cfg *config.Config
	db  *pgxpool.Pool
}

func New(
	ctx context.Context,
	cfg *config.Config,
) (*App, error) {
	pgPool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	return &App{
		cfg: cfg,
		db:  pgPool,
	}, nil
}

func (a *App) Run() error {
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Println("application started")

	<-rootCtx.Done()

	return nil
}

func (a *App) Close() error {
	a.db.Close()

	return nil
}
