package rt

import (
	"fmt"

	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/config"
	"github.com/LeHuuHai/server-management/microservices/pkg/apperr"
	"github.com/LeHuuHai/server-management/microservices/pkg/mq"
	"github.com/segmentio/kafka-go"
)

type App struct {
	Config      *config.Config
	KafkaWriter *kafka.Writer
}

func NewApp() (*App, error) {
	// load config
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrAppBuild, err)
	}

	kw := mq.NewWriter(cfg.KafkaConfig, mq.WithRequiredAcks(0))

	return &App{
		Config:      cfg,
		KafkaWriter: kw,
	}, nil
}
