package main

import (
	"log"
	"log/slog"
	"net"
	"strconv"

	"context"

	serverpb "github.com/LeHuuHai/server-management/microservices/pkg/pb/server"
	auth "github.com/LeHuuHai/server-management/microservices/pkg/auth"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/config"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/infra/kafka"
	pg "github.com/LeHuuHai/server-management/microservices/server-service/internal/infra/postgres"
	rt "github.com/LeHuuHai/server-management/microservices/server-service/internal/infra/runtime"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/infra/service"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/infra/worker"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/rpc"
	"google.golang.org/grpc"
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

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(auth.RoleCheckUnaryGRPCInterceptor(map[string]auth.Scope{
			serverpb.ServerService_CreateServer_FullMethodName:  auth.ScopeServerCreate,
			serverpb.ServerService_UpdateServer_FullMethodName:  auth.ScopeServerUpdate,
			serverpb.ServerService_DeleteServer_FullMethodName:  auth.ScopeServerDelete,
			serverpb.ServerService_ListServers_FullMethodName:   auth.ScopeServerRead,
			serverpb.ServerService_ImportServers_FullMethodName: auth.ScopeServerImport,
			serverpb.ServerService_ExportServers_FullMethodName: auth.ScopeServerExport,
		})),
	)
	serverpb.RegisterServerServiceServer(grpcServer, grpcHandler)

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
