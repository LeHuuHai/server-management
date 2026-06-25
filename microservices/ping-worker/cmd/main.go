package main

import (
	"context"
	"log"
	"sync"

	"github.com/LeHuuHai/server-management/microservices/ping-worker/internal/config"
	"github.com/LeHuuHai/server-management/microservices/ping-worker/internal/infra/kafka"
	rt "github.com/LeHuuHai/server-management/microservices/ping-worker/internal/infra/runtime"
	"github.com/LeHuuHai/server-management/microservices/ping-worker/internal/infra/service"
	"github.com/LeHuuHai/server-management/microservices/ping-worker/internal/infra/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	app, err := rt.NewApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize runtime: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pingRequestConsumer := kafka.NewPingRequestConsumer(app.PingRequestReader)
	pingResponseWriter := kafka.NewPingResponsePublisher(app.PingResponseWriter)
	svc := service.NewPingService()

	workerPool := worker.NewPingWorkerPool(cfg.AppConfig.NumThread, pingRequestConsumer, pingResponseWriter, svc)

	var wg sync.WaitGroup
	wg.Add(1)
	workerPool.Start(ctx, &wg)
	wg.Wait()
}
