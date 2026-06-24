package main

import (
	"log"
	"log/slog"
	"net"
	"strconv"

	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/internal/config"
	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/internal/infra/kafka"
	rt "github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/internal/infra/runtime"
	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/internal/infra/service"
	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/internal/rpc"
	heartbeatpb "github.com/LeHuuHai/server-management/microservices/pkg/pb/heartbeat"
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

	publisher := kafka.NewKafkaPublisher(app.KafkaWriter)
	gwSvc := service.NewGwService(publisher)
	grpcHandler := rpc.NewHeartbeatHandler(gwSvc)

	grpcServer := grpc.NewServer()
	heartbeatpb.RegisterHeartbeatServiceServer(grpcServer, grpcHandler)

	addr := net.JoinHostPort(cfg.AppConfig.Host, strconv.Itoa(cfg.AppConfig.Port))
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	slog.Info("Starting HeartbeatGateway Service gRPC Server", "addr", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
