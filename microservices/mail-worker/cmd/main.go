package main

import (
	"context"
	"log"
	"sync"

	"github.com/LeHuuHai/server-management/microservices/mail-worker/internal/config"
	"github.com/LeHuuHai/server-management/microservices/mail-worker/internal/infra/kafka"
	"github.com/LeHuuHai/server-management/microservices/mail-worker/internal/infra/mail"
	rt "github.com/LeHuuHai/server-management/microservices/mail-worker/internal/infra/runtime"
	"github.com/LeHuuHai/server-management/microservices/mail-worker/internal/infra/service"
	"github.com/LeHuuHai/server-management/microservices/mail-worker/internal/infra/worker"
	auth "github.com/LeHuuHai/server-management/microservices/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	conn, err := grpc.NewClient(
		app.Config.AppConfig.ReportRepoAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStreamInterceptor(auth.APIKeyBindStreamGRPCInterceptor(cfg.AppConfig.InternalAPIKey)),
	)
	if err != nil {
		log.Fatalf("Failed to connect to report-repo: %v", err)
	}
	defer conn.Close()

	// Dependency Injection Wiring
	mailConsumer := kafka.NewMailConsumer(app.MailReader)
	gomailSender := mail.NewGomailSender(app.GomailDialer, cfg.SenderConfig.From)
	downloadService := service.NewDownloadService(conn)

	mailWorker := worker.NewMailWorker(mailConsumer, gomailSender, downloadService)

	var wg sync.WaitGroup
	wg.Add(1)
	go mailWorker.Start(ctx, &wg)
	wg.Wait()
}
