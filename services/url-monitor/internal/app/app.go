package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"
	"url-monitor/internal/check"
	"url-monitor/internal/config"
	apphttp "url-monitor/internal/http"
	"url-monitor/internal/metrics"
	"url-monitor/internal/monitor"
	"url-monitor/internal/pool"
	"url-monitor/internal/storage/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
)

type Runner interface {
	Run(ctx context.Context) error
}

type App struct {
	cfg       *config.Config
	server    *http.Server
	db        *pgxpool.Pool
	scheduler *monitor.Scheduler
	runners   []Runner
}

func New(
	ctx context.Context,
	addr string,
	cfg *config.Config,
) (*App, error) {
	pgPool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	m := metrics.NewMetrics(prometheus.DefaultRegisterer)

	repo := postgres.NewMonitorRepository(pgPool)
	monitorService := monitor.NewMonitorService(repo)

	checkService := check.NewCheckStoreService(repo)
	checker := &check.CheckRunner{}
	checkProcessor := check.NewCheckProcessor(checker, checkService, m)

	monitorDispatcher := pool.NewWorkerPool[monitor.Monitor](checkProcessor, cfg.MonitorCheckWorkersCount, cfg.MonitorCheckQueueSize, m)
	scheduler := monitor.NewScheduler(repo, monitorDispatcher, time.Duration(cfg.SchedulerTimeInterval)*time.Second)

	handler := apphttp.NewHandler(monitorService, m)
	router := apphttp.NewRouter(handler)

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	return &App{
		cfg:       cfg,
		server:    server,
		db:        pgPool,
		scheduler: scheduler,
		runners: []Runner{
			monitorDispatcher,
		},
	}, nil
}

func (a *App) Run() error {

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(rootCtx)

	g.Go(func() error {
		return a.scheduler.Run(ctx)
	})

	for _, runner := range a.runners {
		r := runner
		g.Go(func() error {
			return r.Run(ctx)
		})
	}

	g.Go(func() error {
		err := a.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	})

	g.Go(func() error {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := a.server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	})

	return g.Wait()
}

func (a *App) Close(ctx context.Context) error {
	var result error

	if err := a.server.Shutdown(ctx); err != nil {
		result = fmt.Errorf("shutdown server: %w", err)
	}

	a.db.Close()

	return result
}
