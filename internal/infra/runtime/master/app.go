package masterruntime

import (
	masterconfig "github.com/LeHuuHai/server-management/config/master"
	es "github.com/LeHuuHai/server-management/internal/infra/elasticsearch"
	kfk "github.com/LeHuuHai/server-management/internal/infra/kafka"
	pg "github.com/LeHuuHai/server-management/internal/infra/postgres"
	rdb "github.com/LeHuuHai/server-management/internal/infra/redis"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

type App struct {
	Config      *masterconfig.Config
	DB          *gorm.DB
	ESClient    *elasticsearch.Client
	SyncWriter  *kafka.Writer
	AsyncWriter *kafka.Writer
	RdbClient   *redis.Client
}

func NewApp(cfg *masterconfig.Config) (*App, error) {
	// load config
	cfg, err := masterconfig.Load()
	if err != nil {
		panic(err)
	}

	// infra
	db, err := pg.Connect(cfg.DBConfig)
	if err != nil {
		return nil, err
	}
	esclient, err := es.Connect(cfg.ESConfig)
	if err != nil {
		return nil, err
	}
	syncWriter, asyncWriter, err := kfk.ConnectWriter(cfg.KafkaWriterConfig)
	if err != nil {
		return nil, err
	}
	rdbClient, err := rdb.Connect(cfg.RedisConfig)
	if err != nil {
		return nil, err
	}

	return &App{
		Config:      cfg,
		DB:          db,
		ESClient:    esclient,
		SyncWriter:  syncWriter,
		AsyncWriter: asyncWriter,
		RdbClient:   rdbClient,
	}, nil
}
