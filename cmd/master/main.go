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
	"github.com/LeHuuHai/server-management/config"
	"github.com/LeHuuHai/server-management/internal/domain/cache"
	"github.com/LeHuuHai/server-management/internal/handler"
	xlsximport "github.com/LeHuuHai/server-management/internal/infra/file/deserialize"
	xlsxexport "github.com/LeHuuHai/server-management/internal/infra/file/export"
	"github.com/LeHuuHai/server-management/internal/infra/inmem"
	kfk "github.com/LeHuuHai/server-management/internal/infra/kafka"
	pg "github.com/LeHuuHai/server-management/internal/infra/postgres"
	"github.com/LeHuuHai/server-management/internal/infra/runtime"
	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/LeHuuHai/server-management/internal/service"
	"github.com/gin-gonic/gin"
)

func Serve(
	ctx context.Context,
	wg *sync.WaitGroup,
	rt *runtime.App,
	serverService *service.ServerService,
) {
	defer wg.Done()

	//handler
	serverHandler := handler.NewServerHandler(
		serverService,
		xlsxexport.NewServerXLSXExporter(),
		xlsximport.NewServerXLSXImporter(),
	)

	// router
	r := gin.New()
	api.RegisterHandlers(r, serverHandler)
	addr := net.JoinHostPort(
		rt.Config.App.Host,
		strconv.Itoa(rt.Config.App.Port),
	)

	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func CheckServer(
	ctx context.Context,
	wg *sync.WaitGroup,
	rt *runtime.App,
	publishPingService *service.PublishPingService,
	serverMetadataCache cache.ServerMetadataCacheInterface,
) {
	defer wg.Done()
	ticker := time.NewTicker(time.Duration(rt.Config.App.CyclePing) * time.Millisecond)
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
				}
				publishPingService.PublishRequestPing(ctx, reqBytes)
			}
		}
	}
}

func Report(ctx context.Context, wg *sync.WaitGroup, app *runtime.App) {
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
			// agg
			// send res
		}
		timer.Stop()
	}
}

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	rt, err := runtime.New(cfg)
	if err != nil {
		panic(err)
	}

	// domain
	serverRepo := pg.NewServerRepository(rt.DB)
	serverInmemCache := inmem.NewServerInmemCache()
	kfkPublisher := kfk.NewPublisher(rt.SyncWriter)

	// service
	serverService := service.NewServerService(serverRepo, serverInmemCache)
	publishPingService := service.NewPublishPingService(kfkPublisher)

	var wg sync.WaitGroup
	wg.Add(3)
	go Serve(ctx, &wg, rt, serverService)
	go CheckServer(ctx, &wg, rt, publishPingService, serverInmemCache)
	go Report(ctx, &wg, rt)
	wg.Wait()
}
