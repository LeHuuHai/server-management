package main

import (
	"context"
	"log"
	"net"
	"strconv"

	gwapi "github.com/LeHuuHai/server-management/api/gw"
	gwconfig "github.com/LeHuuHai/server-management/config/gw"
	"github.com/LeHuuHai/server-management/internal/handler"
	kfk "github.com/LeHuuHai/server-management/internal/infra/kafka"
	gwruntime "github.com/LeHuuHai/server-management/internal/infra/runtime/gw"
	"github.com/LeHuuHai/server-management/internal/middleware"
	"github.com/LeHuuHai/server-management/internal/service"
	"github.com/gin-gonic/gin"
)

func Serve(
	ctx context.Context,
	rt *gwruntime.App,
	h *handler.GwHandler,
	mw gin.HandlerFunc,
) {
	strictHandler := gwapi.NewStrictHandler(h, nil)

	// router
	r := gin.Default()

	r.Use(mw)
	gwapi.RegisterHandlers(r, strictHandler)

	addr := net.JoinHostPort(
		rt.Config.AppConfig.Host,
		strconv.Itoa(rt.Config.AppConfig.Port),
	)

	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func main() {
	ctx := context.Background()

	cfg, err := gwconfig.Load()
	if err != nil {
		panic(err)
	}

	rt, err := gwruntime.NewApp(cfg)
	if err != nil {
		panic(err)
	}

	// domain, infra
	kfkPublisher := kfk.NewPublisher(rt.AsyncWriter)

	// service
	gwService := service.NewGwService(kfkPublisher, rt.Config.KafkaConfig.Topics["heartbeat"])

	// middleware
	mw := middleware.NewAPIKeyMiddleware(rt.Config.AppConfig.HeartbeatKey)

	// handler
	gwHandler := handler.NewGwHandler(
		gwService,
	)

	Serve(ctx, rt, gwHandler, mw)
}
