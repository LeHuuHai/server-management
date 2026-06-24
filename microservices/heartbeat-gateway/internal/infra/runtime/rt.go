package rt

import (
	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/internal/config"
	"github.com/LeHuuHai/server-management/microservices/pkg/mq"
	"github.com/segmentio/kafka-go"
)

type App struct {
	Config      *config.Config
	KafkaWriter *kafka.Writer
}

func NewApp(cfg *config.Config) (*App, error) {
	kw := mq.NewWriter(cfg.KafkaConfig.Broker, cfg.KafkaConfig.ServerTopic, mq.WithRequiredAcks(0))

	return &App{
		Config:      cfg,
		KafkaWriter: kw,
	}, nil
}
