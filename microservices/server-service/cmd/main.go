package main

import (
	"log"
	"log/slog"
	"net"
	"strconv"

	"github.com/LeHuuHai/server-management/microservices/pkg/pb/server"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/config"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/infra/kafka"
	pg "github.com/LeHuuHai/server-management/microservices/server-service/internal/infra/postgres"
	rt "github.com/LeHuuHai/server-management/microservices/server-service/internal/infra/runtime"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/infra/service"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/infra/worker"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/rpc"
	"google.golang.org/grpc"
	"context"
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

	grpcHandler := rpc.NewServerHandler(serverSvc)

	grpcServer := grpc.NewServer()
	server.RegisterServerServiceServer(grpcServer, grpcHandler)

	addr := net.JoinHostPort(cfg.AppConfig.Host, strconv.Itoa(cfg.AppConfig.Port))
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	outboxPoller := worker.NewOutboxPoller(outboxRepo, eventPublisher)
	go outboxPoller.Start(context.Background())

	slog.Info("Starting Server Service gRPC Server", "addr", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
