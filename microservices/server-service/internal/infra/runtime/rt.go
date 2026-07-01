package rt

import (
	"embed"
	"fmt"
	"log/slog"

	"github.com/LeHuuHai/server-management/microservices/pkg/apperr"
	"github.com/LeHuuHai/server-management/microservices/pkg/db"
	"github.com/LeHuuHai/server-management/microservices/pkg/mq"
	"github.com/LeHuuHai/server-management/microservices/server-service/config"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type App struct {
	Config            *config.Config
	DB                *gorm.DB
	ServerEventWriter *kafka.Writer
}

func NewApp() (*App, error) {
	// load config
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrAppBuild, err)
	}
	slog.Info("runtime: config loaded", "cfg", cfg.LogValue())

	database, err := db.Connect(cfg.DBConfig)
	if err != nil {
		slog.Error("Failed to connect to postgres", "err", err)
		return nil, err
	}
	if err := db.RunMigrations(cfg.DBConfig, migrationsFS, "migrations"); err != nil {
		slog.Error("Failed to run DB migrations", "error", err)
	}

	kw := mq.NewWriter(cfg.KafkaConfig, mq.WithRequiredAcks(-1))

	return &App{
		Config:            cfg,
		DB:                database,
		ServerEventWriter: kw,
	}, nil
}
