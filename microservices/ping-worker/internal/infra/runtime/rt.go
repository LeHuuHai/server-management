package rt

import (
	"github.com/LeHuuHai/server-management/microservices/ping-worker/internal/config"
	"github.com/LeHuuHai/server-management/microservices/pkg/mq"
	"github.com/segmentio/kafka-go"
)

type App struct {
	Config             *config.Config
	PingRequestReader  *kafka.Reader
	PingResponseWriter *kafka.Writer
}

func NewApp(cfg *config.Config) (*App, error) {
	// Initialize Kafka Reader and Writer
	pingReader := mq.NewReader(cfg.PingRequestReaderConfig)
	pingWriter := mq.NewWriter(cfg.PingResponseWriterConfig, mq.WithRequiredAcks(0))

	return &App{
		Config:             cfg,
		PingRequestReader:  pingReader,
		PingResponseWriter: pingWriter,
	}, nil
}
