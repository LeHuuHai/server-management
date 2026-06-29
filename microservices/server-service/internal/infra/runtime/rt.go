package rt

import (
	"embed"
	"log/slog"

	"github.com/LeHuuHai/server-management/microservices/pkg/db"
	"github.com/LeHuuHai/server-management/microservices/pkg/mq"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/config"
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

func NewApp(cfg *config.Config) (*App, error) {
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
