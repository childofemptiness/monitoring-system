package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"url-monitor/internal/app"
	"url-monitor/internal/config"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("warning: .env file not found")
	}

	appCfg, err := config.LoadAppConfig()
	if err != nil {
		log.Fatal(err)
	}

	monitorChecksCfg, err := config.LoadMonitorChecksConfig()
	if err != nil {
		log.Fatal(err)
	}

	outboxEventsCfg, err := config.LoadOutboxEventsConfig()
	if err != nil {
		log.Fatal(err)
	}

	cfg := config.Config{
		AppConfig:           appCfg,
		MonitorChecksConfig: monitorChecksCfg,
		OutboxEventsConfig:  outboxEventsCfg,
	}

	ctx := context.Background()

	application, err := app.New(ctx, ":"+appCfg.AppPort, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		log.Printf("server started on :%s", cfg.AppConfig)

		if err := application.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := application.Close(shutdownCtx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
}
