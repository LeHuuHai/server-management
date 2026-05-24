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
	PingReader  *kafka.Reader
	MailReader  *kafka.Reader
}

func NewApp(cfg *workerconfig.Config) (*App, error) {
	// load config
	cfg, err := workerconfig.Load()
	if err != nil {
		panic(err)
	}

	// infra
	syncWriter, asyncWriter, err := kfk.ConnectWriter(cfg.KafkaConfig)
	if err != nil {
		return nil, err
	}
	pingReader, mailReader, err := kfk.ConnectWorkerReader(cfg.KafkaConfig)
	if err != nil {
		return nil, err
	}

	return &App{
		Config:      cfg,
		SyncWriter:  syncWriter,
		AsyncWriter: asyncWriter,
		PingReader:  pingReader,
		MailReader:  mailReader,
	}, nil
}
