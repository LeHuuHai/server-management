package rt

import (
	"fmt"

	"github.com/LeHuuHai/server-management/microservices/ping-worker/config"
	"github.com/LeHuuHai/server-management/microservices/pkg/apperr"
	"github.com/LeHuuHai/server-management/microservices/pkg/mq"
	"github.com/segmentio/kafka-go"
)

type App struct {
	Config             *config.Config
	PingRequestReader  *kafka.Reader
	PingResponseWriter *kafka.Writer
}

func NewApp() (*App, error) {
	// load config
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrAppBuild, err)
	}

	// Initialize Kafka Reader and Writer
	pingReader := mq.NewReader(cfg.PingRequestReaderConfig)
	pingWriter := mq.NewWriter(cfg.PingResponseWriterConfig, mq.WithRequiredAcks(0))

	return &App{
		Config:             cfg,
		PingRequestReader:  pingReader,
		PingResponseWriter: pingWriter,
	}, nil
}
