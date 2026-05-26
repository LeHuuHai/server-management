package eswriterruntime

import (
	"fmt"

	eswriterconfig "github.com/LeHuuHai/server-management/config/eswriter"
	apperr "github.com/LeHuuHai/server-management/internal/error"
	es "github.com/LeHuuHai/server-management/internal/infra/elasticsearch"
	kfk "github.com/LeHuuHai/server-management/internal/infra/kafka"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/segmentio/kafka-go"
)

type App struct {
	Config          *eswriterconfig.Config
	ESClient        *elasticsearch.Client
	HeartBeatReader *kafka.Reader
}

func NewApp(cfg *eswriterconfig.Config) (*App, error) {
	// load config
	cfg, err := eswriterconfig.Load()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrAppBuild, err)
	}

	// infra
	esclient, err := es.Connect(cfg.ESConfig)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrAppBuild, err)
	}
	heartbeatReader, err := kfk.ConnectHeartbeatReader(cfg.KafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrAppBuild, err)
	}

	return &App{
		Config:          cfg,
		ESClient:        esclient,
		HeartBeatReader: heartbeatReader,
	}, nil
}
