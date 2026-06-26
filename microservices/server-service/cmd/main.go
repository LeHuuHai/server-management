package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/LeHuuHai/server-management/microservices/server-service/api"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/config"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/infra/handler"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/infra/kafka"
	pg "github.com/LeHuuHai/server-management/microservices/server-service/internal/infra/postgres"
	rt "github.com/LeHuuHai/server-management/microservices/server-service/internal/infra/runtime"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/infra/service"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/infra/worker"
	"github.com/gin-gonic/gin"
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

	serverRepo := pg.NewServerRepository(app.DB)
	txManager := pg.NewTxManager(app.DB)
	outboxRepo := pg.NewOutboxRepository(app.DB)

	eventPublisher := kafka.NewServerEventPublisher(app.ServerEventWriter)

	serverSvc := service.NewServerService(txManager, serverRepo, outboxRepo)

	serverHandler := handler.NewServerRestHandler(serverSvc)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// In the future, you could add authorization middleware here to check X-User-Role header
	strictHandler := api.NewStrictHandler(serverHandler, nil)
	api.RegisterHandlers(router, strictHandler)

	addr := net.JoinHostPort(cfg.AppConfig.Host, strconv.Itoa(cfg.AppConfig.Port))
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	outboxPoller := worker.NewOutboxPoller(outboxRepo, eventPublisher)
	go outboxPoller.Start(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("Starting Server Service REST Server", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to listen and serve: %v", err)
		}
	}()

	<-sigChan
	slog.Info("Shutdown signal received, shutting down HTTP server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", slog.Any("error", err))
	}

	slog.Info("Server Service HTTP server stopped gracefully")
}
