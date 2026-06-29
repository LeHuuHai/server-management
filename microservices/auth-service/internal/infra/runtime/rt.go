package rt

import (
	"fmt"
	"log/slog"

	"github.com/LeHuuHai/server-management/microservices/auth-service/internal/config"
	"github.com/LeHuuHai/server-management/microservices/pkg/apperr"
	"github.com/LeHuuHai/server-management/microservices/pkg/cache"
	"github.com/LeHuuHai/server-management/microservices/pkg/db"
	jwtprovider "github.com/LeHuuHai/server-management/microservices/pkg/jwt"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type App struct {
	Config      *config.Config
	JWTProvider *jwtprovider.JWTProvider
	DB          *gorm.DB
	RdbClient   *redis.Client
}

func NewApp() (*App, error) {
	// load config
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrAppBuild, err)
	}

	// infra
	jwtProvider := jwtprovider.NewJWTProvider(cfg.JWTConfig)
	db, err := db.Connect(cfg.DBConfig)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrAppBuild, err)
	}
	rdbClient, err := cache.Connect(cfg.RedisConfig)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrAppBuild, err)
	}

	app := App{
		Config:      cfg,
		JWTProvider: jwtProvider,
		DB:          db,
		RdbClient:   rdbClient,
	}
	slog.Info("App initialized successfully", "app", app)
	return &app, nil
}
