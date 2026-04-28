package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/childofemptiness/alert-service/internal/app"
	"github.com/childofemptiness/alert-service/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("warning: .env file not found")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	application, err := app.New(ctx, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := application.Run(); err != nil {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := application.Close(); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
}
