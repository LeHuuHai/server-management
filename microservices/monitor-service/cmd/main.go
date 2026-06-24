package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/config"
	domainservice "github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/service"
	esagg "github.com/LeHuuHai/server-management/microservices/monitor-service/internal/infra/elasticsearch"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/infra/kafka"
	pg "github.com/LeHuuHai/server-management/microservices/monitor-service/internal/infra/postgres"
	rdb "github.com/LeHuuHai/server-management/microservices/monitor-service/internal/infra/redis"
	rt "github.com/LeHuuHai/server-management/microservices/monitor-service/internal/infra/runtime"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/infra/service"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/infra/worker"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/model"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/rpc"
	pb "github.com/LeHuuHai/server-management/microservices/pkg/pb/monitor"
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

	pingRequestPublisher := kafka.NewPingRequestPublisher(app.PingRequestWriter)
	mailPublisher := kafka.NewMailPublisher(app.MailWriter)
	serverEventsConsumer := kafka.NewLifecycleConsumer(app.ServerEventsReader)
	heartbeatConsumer := kafka.NewHeartbeatConsumer(app.HeartbeatReader)
	pingResponseConsumer := kafka.NewPingResponseConsumer(app.PingResponseReader)

	// Channels for batch processing
	pgChan := make(chan model.LiveStatus, 4000)
	esChan := make(chan model.StatusLog, 4000)

	// Repositories & Clients
	liveStatusRepo := pg.NewLiveStatusRepository(app.DB)
	monitoredServerRepo := pg.NewMonitoredServerRepository(app.DB)
	esWriter := esagg.NewESWriter[model.StatusLog](app.ESClient, cfg.ESConfig.Index)
	esAggregator := esagg.NewESAggregator(app.ESClient, cfg.ESConfig.Index)
	redisCache := rdb.NewDailyReportRedisCache(app.RedisClient)

	// Core Services
	monitorSvc := service.NewMonitorService(monitoredServerRepo, liveStatusRepo, pgChan, esChan)
	reportSvc := service.NewReportService(esAggregator, redisCache, mailPublisher)
	batchPG := service.NewBatchPGService(pgChan, 2000, time.Second, liveStatusRepo)
	batchES := service.NewBatchESService(esChan, 2000, time.Second, esWriter)

	// Background Workers
	serverLifecycleHandler := worker.NewLifecycleConsumer(serverEventsConsumer, monitorSvc)
	heartbeatHandler := worker.NewHeartbeatConsumer(heartbeatConsumer, monitorSvc)
	pingResultHandler := worker.NewPingResultConsumer(pingResponseConsumer, monitorSvc)
	activeChecker := worker.NewActiveChecker(
		monitoredServerRepo,
		liveStatusRepo,
		pingRequestPublisher,
		cfg.AppConfig.CyclePing,
		cfg.AppConfig.HeartbeatTimeout,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// Start Batch services
	wg.Add(2)
	go func() {
		defer wg.Done()
		batchPG.Run(ctx)
	}()
	go func() {
		defer wg.Done()
		batchES.Run(ctx)
	}()

	// Start Background Consumers, Checker & Daily Report
	wg.Add(5)
	go func() {
		defer wg.Done()
		serverLifecycleHandler.Start(ctx)
	}()
	go func() {
		defer wg.Done()
		heartbeatHandler.Start(ctx)
	}()
	go func() {
		defer wg.Done()
		pingResultHandler.Start(ctx)
	}()
	go func() {
		defer wg.Done()
		activeChecker.Start(ctx)
	}()
	go func() {
		defer wg.Done()
		Report(ctx, cfg.AppConfig.AdMail, reportSvc)
	}()

	// Start gRPC Server
	grpcHandler := rpc.NewMonitorHandler()
	grpcServer := grpc.NewServer()
	pb.RegisterMonitorServiceServer(grpcServer, grpcHandler)

	addr := net.JoinHostPort(cfg.AppConfig.Host, strconv.Itoa(cfg.AppConfig.Port))
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on tcp: %v", err)
	}

	slog.Info("Starting Monitor Service gRPC Server", "addr", addr)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Wait for system exit / contexts
	wg.Wait()
}

func Report(
	ctx context.Context,
	adMail string,
	reportSvc domainservice.ReportServiceInterface,
) {
	for {
		now := time.Now()
		today := time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			0, 0, 0, 0,
			now.Location(),
		)
		tomorrow := today.Add(24 * time.Hour)
		timer := time.NewTimer(tomorrow.Sub(now))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			if adMail == "" {
				slog.Warn("Daily report email (AD_MAIL) is not set, skipping report generation")
				continue
			}
			request := model.GenServerReportRequest{
				From:      time.Now().Add(-24 * time.Hour),
				To:        time.Now(),
				Receivers: []string{adMail},
			}
			err := reportSvc.ReportServer(ctx, request)
			if err != nil {
				slog.Warn("Report service failed", slog.Any("request", request), slog.Any("err", err))
				continue
			}
		}
		timer.Stop()
	}
}
