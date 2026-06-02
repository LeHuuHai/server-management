package gwruntime

import (
	"fmt"

	gwconfig "github.com/LeHuuHai/server-management/config/gw"
	apperr "github.com/LeHuuHai/server-management/internal/error"
	kfk "github.com/LeHuuHai/server-management/internal/infra/kafka"
	"github.com/segmentio/kafka-go"
)

type App struct {
	Config      *gwconfig.Config
	SyncWriter  *kafka.Writer
	AsyncWriter *kafka.Writer
}

func NewApp(cfg *gwconfig.Config) (*App, error) {
	// load config
	cfg, err := gwconfig.Load()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrAppBuild, err)
	}

	// infra
	syncWriter, asyncWriter, err := kfk.ConnectWriter(cfg.KafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrAppBuild, err)
	}

	return &App{
		Config:      cfg,
		SyncWriter:  syncWriter,
		AsyncWriter: asyncWriter,
	}, nil
}
