package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/LeHuuHai/server-management/api"
	masterconfig "github.com/LeHuuHai/server-management/config/master"
	"github.com/LeHuuHai/server-management/internal/domain/cache"
	"github.com/LeHuuHai/server-management/internal/domain/mq"
	"github.com/LeHuuHai/server-management/internal/handler"
	es "github.com/LeHuuHai/server-management/internal/infra/elasticsearch"
	xlsximport "github.com/LeHuuHai/server-management/internal/infra/file/deserialize"
	xlsxexport "github.com/LeHuuHai/server-management/internal/infra/file/export"
	"github.com/LeHuuHai/server-management/internal/infra/inmem"
	jwtprovider "github.com/LeHuuHai/server-management/internal/infra/jwt"
	kfk "github.com/LeHuuHai/server-management/internal/infra/kafka"
	pg "github.com/LeHuuHai/server-management/internal/infra/postgres"
	rdb "github.com/LeHuuHai/server-management/internal/infra/redis"
	masterruntime "github.com/LeHuuHai/server-management/internal/infra/runtime/master"
	"github.com/LeHuuHai/server-management/internal/middleware"
	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/LeHuuHai/server-management/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Serve(
	ctx context.Context,
	wg *sync.WaitGroup,
	rt *masterruntime.App,
	h *handler.Handler,
	mw api.StrictMiddlewareFunc,
) {
	defer wg.Done()

	strictHandler := api.NewStrictHandler(h, []api.StrictMiddlewareFunc{mw})

	// router
	r := gin.Default()

	// cors
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:8081",
		},

		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},

		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"X-API-Key",
		},

		ExposeHeaders: []string{
			"Content-Length",
		},

		AllowCredentials: true,
	}))

	api.RegisterHandlers(r, strictHandler)

	addr := net.JoinHostPort(
		rt.Config.AppConfig.Host,
		strconv.Itoa(rt.Config.AppConfig.Port),
	)

	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func CheckServer(
	ctx context.Context,
	wg *sync.WaitGroup,
	rt *masterruntime.App,
	publisher mq.Publisher,
	serverMetadataCache cache.ServerMetadataCacheInterface,
) {
	defer wg.Done()
	ticker := time.NewTicker(time.Duration(rt.Config.AppConfig.CyclePing) * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			start := time.Now()
			servers := serverMetadataCache.List(ctx)
			cnt := 0
			for _, item := range servers {
				if item.LastHeartbeatAt != nil && time.Since(*item.LastHeartbeatAt) < time.Duration(rt.Config.AppConfig.HeartbeatTimeout)*time.Millisecond {
					continue
				}
				cnt++
				req := model.RequestPing{
					ServerID:   item.ServerID,
					ServerName: item.ServerName,
					IP:         item.IPv4,
				}
				reqBytes, err := json.Marshal(req)
				if err != nil {
					log.Println(err.Error())
					continue
				}
				msg := mq.Message{
					Topic: rt.Config.KafkaConfig.Topics["ping"],
					Value: reqBytes,
				}
				err = publisher.Publish(ctx, msg)
				if err != nil {
					log.Println(err.Error())
					continue
				}
			}
			elapse := time.Since(start)
			log.Printf("publish %d servers in %v", cnt, elapse)
		}
	}
}

func Report(
	ctx context.Context,
	wg *sync.WaitGroup,
	rt *masterruntime.App,
	reportServerService *service.ReportServerService,
) {
	defer wg.Done()
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
			request := model.GenServerReportRequest{
				From:      time.Now().Add(-24 * time.Hour),
				To:        time.Now(),
				Receivers: []string{rt.Config.AppConfig.AdMail},
			}
			err := reportServerService.ReportServer(ctx, request)
			if err != nil {
				log.Println(err.Error())
				continue
			}
		}
		timer.Stop()
	}
}

func ReadTopic(
	ctx context.Context,
	wg *sync.WaitGroup,
	consumer *kfk.KfkConsumer,
	ch chan<- model.ServerMetadata,
) {
	defer wg.Done()
	for {
		msg, err := consumer.Read(ctx)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				continue
			}
		}
		consumer.Commit(ctx, msg)
		var res model.Heartbeat
		err = json.Unmarshal(msg.Value, &res)
		if err != nil {
			continue
		}
		s := model.ServerMetadata{
			ServerID:        res.ServerID,
			LastHeartbeatAt: &res.Timestamp,
		}
		select {
		case <-ctx.Done():
			return

		case ch <- s:
		}
	}
}

func ListenHeartbeat(
	ctx context.Context,
	wg *sync.WaitGroup,
	rt *masterruntime.App,
	serverMetadataCache cache.ServerMetadataCacheInterface,
) {
	defer wg.Done()
	var heartbeatWg sync.WaitGroup
	consumer := kfk.NewConsumer(rt.HeartbeatReader)
	ch := make(chan model.ServerMetadata, 4000)
	batchService := service.NewBatchServerMetadataService(
		ch,
		2000,
		time.Second,
		func(items []model.ServerMetadata) error {
			serverMetadataCache.BatchUpdateHeartbeat(ctx, items)
			return nil
		},
	)
	heartbeatWg.Add(2)
	go ReadTopic(ctx, &heartbeatWg, consumer, ch)
	go func() {
		defer heartbeatWg.Done()
		batchService.Run(ctx)
	}()
	heartbeatWg.Wait()
}

func main() {
	ctx := context.Background()

	cfg, err := masterconfig.Load()
	if err != nil {
		panic(err)
	}

	rt, err := masterruntime.NewApp(cfg)
	if err != nil {
		panic(err)
	}

	// domain, infra
	serverRepo := pg.NewServerRepository(rt.DB)
	kfkPublisher := kfk.NewPublisher(rt.AsyncWriter)
	dailyReportRedisCache := rdb.NewDailyReportRedisCache(rt.RdbClient)
	esAggregator := es.NewESAggregator(rt.ESClient, rt.Config.ESConfig.Index)
	esCachedAggregator := es.NewCachedAggregator(esAggregator, dailyReportRedisCache)
	reportServerXLSXExporter := xlsxexport.NewReportServerXLSXExporter()
	jwtProvider := jwtprovider.NewJWTProvider(rt.Config.JWTConfig)
	tokenBlocklistRedis := jwtprovider.NewTokenBlocklistRedis(rt.RdbClient)
	accountRepo := pg.NewAccountRepository(rt.DB)

	// service
	serverInmemCache, err := inmem.NewServerInmemCache(ctx, serverRepo)
	if err != nil {
		panic(err)
	}
	serverService := service.NewServerService(serverRepo, serverInmemCache)
	reportServerService := service.NewReportServerService(esCachedAggregator, reportServerXLSXExporter, kfkPublisher, rt.Config.KafkaConfig.Topics["mail"])
	authService := service.NewAuthService(jwtProvider, tokenBlocklistRedis, accountRepo)

	// middleware
	mw := middleware.NewAuthStrictMiddleware(jwtProvider, tokenBlocklistRedis, rt.Config.AppConfig.ReportKey)

	// handler
	serverHandler := handler.NewServerHandler(
		serverService,
		reportServerService,
		xlsxexport.NewServerXLSXExporter(),
		xlsximport.NewServerXLSXImporter(),
	)
	authHandler := handler.NewAuthHandler(authService)
	h := handler.NewHandler(serverHandler, authHandler)

	var wg sync.WaitGroup
	wg.Add(4)
	go Serve(ctx, &wg, rt, h, mw)
	go CheckServer(ctx, &wg, rt, kfkPublisher, serverInmemCache)
	go Report(ctx, &wg, rt, reportServerService)
	go ListenHeartbeat(ctx, &wg, rt, serverInmemCache)
	wg.Wait()
}
