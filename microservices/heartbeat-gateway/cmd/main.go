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

	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/api"
	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/internal/config"
	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/internal/handler"
	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/internal/infra/kafka"
	rt "github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/internal/infra/runtime"
	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/internal/infra/service"
	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/internal/middleware"
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

	publisher := kafka.NewKafkaPublisher(app.KafkaWriter)
	gwSvc := service.NewGwService(publisher)
	httpHandler := handler.NewHeartbeatHandler(gwSvc)

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Register API Key Middleware and Strict API Handlers
	router.Use(middleware.NewAPIKeyMiddleware(cfg.AppConfig.HeartbeatKey))
	
	strictHandler := api.NewStrictHandler(httpHandler, nil)
	api.RegisterHandlers(router, strictHandler)

	addr := net.JoinHostPort(cfg.AppConfig.Host, strconv.Itoa(cfg.AppConfig.Port))
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Capture OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("Starting HeartbeatGateway Service HTTP Server", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to listen and serve: %v", err)
		}
	}()

	// Block until a shutdown signal is received
	<-sigChan
	slog.Info("Shutdown signal received, shutting down HTTP server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", slog.Any("error", err))
	}

	slog.Info("HeartbeatGateway HTTP server stopped gracefully")
}
