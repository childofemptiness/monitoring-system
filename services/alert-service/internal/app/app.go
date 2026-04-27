package app

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/childofemptiness/alert-service/internal/config"
	inbox "github.com/childofemptiness/alert-service/internal/events"
	"github.com/childofemptiness/alert-service/internal/storage/postgres"
	"github.com/childofemptiness/monitoring-system/contracts/events"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

type Consumer interface {
	Consume(ctx context.Context, handler func(ctx context.Context, event events.EventEnvelope) error) error
}

type InboxService interface {
	HandleEvent(ctx context.Context, event events.EventEnvelope) error
}

type App struct {
	cfg      *config.Config
	db       *pgxpool.Pool
	svc      InboxService
	consumer Consumer
}

func New(
	ctx context.Context,
	cfg *config.Config,
) (*App, error) {
	pgPool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	repo := postgres.NewInboxRepository(pgPool)
	svc := inbox.NewInboxService(repo)

	return &App{
		cfg: cfg,
		db:  pgPool,
		svc: &svc,
	}, nil
}

func (a *App) Run() error {
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(rootCtx)

	log.Println("application started")

	g.Go(func() error {
		return a.consumer.Consume(ctx, a.svc.HandleEvent)
	})

	g.Go(func() error {
		<-rootCtx.Done()

		log.Println("application stopped")

		return nil
	})

	return g.Wait()
}

func (a *App) Close() error {
	a.db.Close()

	return nil
}
