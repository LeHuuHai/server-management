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
	"github.com/LeHuuHai/server-management/internal/handler"
	es "github.com/LeHuuHai/server-management/internal/infra/elasticsearch"
	xlsximport "github.com/LeHuuHai/server-management/internal/infra/file/deserialize"
	xlsxexport "github.com/LeHuuHai/server-management/internal/infra/file/export"
	"github.com/LeHuuHai/server-management/internal/infra/inmem"
	kfk "github.com/LeHuuHai/server-management/internal/infra/kafka"
	pg "github.com/LeHuuHai/server-management/internal/infra/postgres"
	masterruntime "github.com/LeHuuHai/server-management/internal/infra/runtime/master"
	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/LeHuuHai/server-management/internal/service"
	"github.com/gin-gonic/gin"
)

func Serve(
	ctx context.Context,
	wg *sync.WaitGroup,
	rt *masterruntime.App,
	serverService *service.ServerService,
	reportServerService *service.ReportServerService,
) {
	defer wg.Done()

	//handler
	serverHandler := handler.NewServerHandler(
		serverService,
		reportServerService,
		xlsxexport.NewServerXLSXExporter(),
		xlsximport.NewServerXLSXImporter(),
	)

	// router
	r := gin.New()
	api.RegisterHandlers(r, serverHandler)
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
	publishPingService *service.PublishService,
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
			servers := serverMetadataCache.List(ctx)
			for _, items := range servers {
				req := model.RequestPing{
					ServerID:   items.ServerID,
					ServerName: items.ServerName,
					IP:         items.IPv4,
				}
				reqBytes, err := json.Marshal(req)
				if err != nil {
					log.Println(err.Error())
					continue
				}
				err = publishPingService.Publish(ctx, "ping", reqBytes)
				if err != nil {
					log.Println(err.Error())
					continue
				}
			}
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
	serverInmemCache := inmem.NewServerInmemCache()
	kfkPublisher := kfk.NewPublisher(rt.SyncWriter)
	esAggregator := es.NewESAggregator(rt.ESClient)
	reportServerXLSXExporter := xlsxexport.NewReportServerXLSXExporter()

	// service
	serverService := service.NewServerService(serverRepo, serverInmemCache)
	reportServerService := service.NewReportServerService(esAggregator, reportServerXLSXExporter, kfkPublisher)
	publishService := service.NewPublishService(kfkPublisher)

	var wg sync.WaitGroup
	wg.Add(3)
	go Serve(ctx, &wg, rt, serverService, reportServerService)
	go CheckServer(ctx, &wg, rt, publishService, serverInmemCache)
	go Report(ctx, &wg, rt, reportServerService)
	wg.Wait()
}
