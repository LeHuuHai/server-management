package pgwriterruntime

import (
	"fmt"

	pgwriterconfig "github.com/LeHuuHai/server-management/config/pgwriter"
	apperr "github.com/LeHuuHai/server-management/internal/error"
	kfk "github.com/LeHuuHai/server-management/internal/infra/kafka"
	pg "github.com/LeHuuHai/server-management/internal/infra/postgres"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

type App struct {
	Config          *pgwriterconfig.Config
	DB              *gorm.DB
	PingResReader   *kafka.Reader
	HeartbeatReader *kafka.Reader
}

func NewApp(cfg *pgwriterconfig.Config) (*App, error) {
	// load config
	cfg, err := pgwriterconfig.Load()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrAppBuild, err)
	}

	// infra
	pingResReader, err := kfk.ConnectPingResReader(cfg.KafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrAppBuild, err)
	}
	heartbeatReader, err := kfk.ConnectHeartbeatReader(cfg.KafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrAppBuild, err)
	}
	db, err := pg.Connect(cfg.DBConfig)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrAppBuild, err)
	}

	return &App{
		Config:          cfg,
		DB:              db,
		PingResReader:   pingResReader,
		HeartbeatReader: heartbeatReader,
	}, nil
}
