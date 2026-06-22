package main

import (
	"log"
	"log/slog"
	"net"
	"strconv"

	"github.com/LeHuuHai/server-management/microservices/auth-service/api"
	"github.com/LeHuuHai/server-management/microservices/auth-service/internal/config"
	"github.com/LeHuuHai/server-management/microservices/auth-service/internal/handler"
	pg "github.com/LeHuuHai/server-management/microservices/auth-service/internal/infra/postgres"
	rdb "github.com/LeHuuHai/server-management/microservices/auth-service/internal/infra/redis"
	rt "github.com/LeHuuHai/server-management/microservices/auth-service/internal/infra/runtime"
	"github.com/LeHuuHai/server-management/microservices/auth-service/internal/service"
	jwtprovider "github.com/LeHuuHai/server-management/microservices/pkg/jwt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	rt, err := rt.NewApp(cfg)
	if err != nil {
		panic(err)
	}

	jwtProvider := jwtprovider.NewJWTProvider(rt.Config.JWTConfig)
	tokenBlocklistRedis := rdb.NewTokenBlocklistRedis(rt.RdbClient)
	accountRepo := pg.NewAccountRepository(rt.DB)

	authService := service.NewAuthService(jwtProvider, tokenBlocklistRedis, accountRepo)
	authHandler := handler.NewAuthHandler(authService)

	strictHandler := api.NewStrictHandler(authHandler, []api.StrictMiddlewareFunc{})

	corsConfig := cors.Config{
		AllowOrigins:     rt.Config.AppConfig.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-API-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}

	r := gin.Default()
	r.Use(cors.New(corsConfig))

	api.RegisterHandlers(r, strictHandler)

	addr := net.JoinHostPort(rt.Config.AppConfig.Host, strconv.Itoa(rt.Config.AppConfig.Port))
	slog.Info("Starting Auth Service", "addr", addr)

	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
