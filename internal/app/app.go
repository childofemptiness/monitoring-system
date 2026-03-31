package app

import (
	"context"
	"fmt"
	"net/http"
	"time"
	check2 "url-monitor/internal/check"
	"url-monitor/internal/config"
	apphttp "url-monitor/internal/http"
	"url-monitor/internal/metrics"
	"url-monitor/internal/monitor"
	"url-monitor/internal/storage/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

type App struct {
	cfg        *config.Config
	server     *http.Server
	db         *pgxpool.Pool
	scheduler  *monitor.Scheduler
	workerPool *monitor.WorkerPool
}

func New(
	ctx context.Context,
	addr string,
	cfg *config.Config,
) (*App, error) {
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	m := metrics.NewMetrics(prometheus.DefaultRegisterer)

	repo := postgres.NewMonitorRepository(pool)
	monitorService := monitor.NewMonitorService(repo)
	checkService := check2.NewCheckStoreService(repo)
	checker := &check2.CheckRunner{}
	processor := check2.NewCheckProcessor(checker, checkService, m)
	workerPool := monitor.NewWorkerPool(processor, cfg.MonitorCheckWorkersCount, cfg.MonitorCheckQueueSize, m)
	scheduler := monitor.NewScheduler(repo, workerPool, time.Duration(cfg.SchedulerTimeInterval)*time.Second)
	handler := apphttp.NewHandler(monitorService, m)
	router := apphttp.NewRouter(handler)

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	return &App{
		cfg:        cfg,
		server:     server,
		db:         pool,
		scheduler:  scheduler,
		workerPool: workerPool,
	}, nil
}

func (a *App) Run() error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 2)

	go func(ctx context.Context) {
		if err := a.scheduler.Run(ctx); err != nil {
			errCh <- err
		}
	}(ctx)

	go func(ctx context.Context) {
		a.workerPool.Run(ctx)
	}(ctx)

	go func() {
		if err := a.server.ListenAndServe(); err != nil {
			errCh <- err
		}

		errCh <- nil
	}()

	err := <-errCh
	defer close(errCh)

	return err
}

func (a *App) Close(ctx context.Context) error {
	var result error

	if err := a.server.Shutdown(ctx); err != nil {
		result = fmt.Errorf("shutdown server: %w", err)
	}

	a.db.Close()

	return result
}
