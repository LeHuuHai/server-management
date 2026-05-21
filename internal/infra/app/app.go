package app

import (
	"github.com/LeHuuHai/server-management/config"
	database "github.com/LeHuuHai/server-management/internal/infra/db"
	es "github.com/LeHuuHai/server-management/internal/infra/elasticsearch"
	kfk "github.com/LeHuuHai/server-management/internal/infra/kafka"
	rdb "github.com/LeHuuHai/server-management/internal/infra/redis"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

type App struct {
	DB          *gorm.DB
	ESClient    *elasticsearch.Client
	SyncWriter  *kafka.Writer
	AsyncWriter *kafka.Writer
	RdbClient   *redis.Client
}

func New(cfg *config.Config) (*App, error) {
	// load config
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// infra
	db, err := database.Connect(cfg)
	if err != nil {
		return nil, err
	}
	esclient, err := es.Connect(cfg)
	if err != nil {
		return nil, err
	}
	syncWriter, asyncWriter, err := kfk.Connect(cfg)
	if err != nil {
		return nil, err
	}
	rdbClient, err := rdb.Connect(cfg)
	if err != nil {
		return nil, err
	}

	return &App{
		DB:          db,
		ESClient:    esclient,
		SyncWriter:  syncWriter,
		AsyncWriter: asyncWriter,
		RdbClient:   rdbClient,
	}, nil
}
