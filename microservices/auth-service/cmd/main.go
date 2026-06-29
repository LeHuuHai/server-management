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

	"github.com/LeHuuHai/server-management/microservices/auth-service/api"
	"github.com/LeHuuHai/server-management/microservices/auth-service/internal/infra/handler"
	pg "github.com/LeHuuHai/server-management/microservices/auth-service/internal/infra/postgres"
	rdb "github.com/LeHuuHai/server-management/microservices/auth-service/internal/infra/redis"
	rt "github.com/LeHuuHai/server-management/microservices/auth-service/internal/infra/runtime"
	"github.com/LeHuuHai/server-management/microservices/auth-service/internal/infra/service"
	"github.com/gin-gonic/gin"
)

func main() {
	app, err := rt.NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize runtime: %v", err)
	}

	blocklist := rdb.NewTokenBlocklistRedis(app.RdbClient)
	repo := pg.NewAccountRepository(app.DB)
	authService := service.NewAuthService(
		app.JWTProvider,
		blocklist,
		repo,
	)

	authHandler := handler.NewAuthRestHandler(authService)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	strictHandler := api.NewStrictHandler(authHandler, nil)
	api.RegisterHandlers(router, strictHandler)

	addr := net.JoinHostPort(app.Config.AppConfig.Host, strconv.Itoa(app.Config.AppConfig.Port))
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("Starting Auth Service REST Server", "addr", addr)
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

	slog.Info("Auth Service HTTP server stopped gracefully")
}
