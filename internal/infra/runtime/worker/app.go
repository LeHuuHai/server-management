package workerruntime

import (
	workerconfig "github.com/LeHuuHai/server-management/config/worker"
	kfk "github.com/LeHuuHai/server-management/internal/infra/kafka"
	"github.com/segmentio/kafka-go"
)

type App struct {
	Config      *workerconfig.Config
	SyncWriter  *kafka.Writer
	AsyncWriter *kafka.Writer
	Reader      *kafka.Reader
}

func NewApp(cfg *workerconfig.Config) (*App, error) {
	// load config
	cfg, err := workerconfig.Load()
	if err != nil {
		panic(err)
	}

	// infra
	syncWriter, asyncWriter, err := kfk.ConnectWriter(cfg.KafkaWriterConfig)
	if err != nil {
		return nil, err
	}
	reader, err := kfk.ConnectReader(cfg.KafkaReaderConfig)
	if err != nil {
		return nil, err
	}

	return &App{
		Config:      cfg,
		SyncWriter:  syncWriter,
		AsyncWriter: asyncWriter,
		Reader:      reader,
	}, nil
}
