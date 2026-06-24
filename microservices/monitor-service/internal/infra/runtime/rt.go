package runtime

import (
	"log/slog"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/config"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/model"
	"github.com/LeHuuHai/server-management/microservices/pkg/db"
	"github.com/LeHuuHai/server-management/microservices/pkg/es"
	pkgmq "github.com/LeHuuHai/server-management/microservices/pkg/mq"
	"github.com/LeHuuHai/server-management/microservices/pkg/rdb"
	esclient "github.com/elastic/go-elasticsearch/v8"
	redisclient "github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

type App struct {
	Config             *config.Config
	DB                 *gorm.DB
	RedisClient        *redisclient.Client
	ESClient           *esclient.Client
	ServerEventsReader *kafka.Reader
	HeartbeatReader    *kafka.Reader
	PingResponseReader *kafka.Reader
	PingRequestWriter  *kafka.Writer
	MailWriter         *kafka.Writer
}

func NewApp(cfg *config.Config) (*App, error) {
	database, err := db.Connect(cfg.DBConfig)
	if err != nil {
		slog.Error("Failed to connect to postgres in monitor-service", "err", err)
		return nil, err
	}

	// AutoMigrate local server status and metadata tables
	err = database.AutoMigrate(&model.MonitoredServer{}, &model.LiveStatus{})
	if err != nil {
		slog.Error("Failed to auto migrate database schemas", "err", err)
		return nil, err
	}

	return NewAppWithDB(cfg, database)
}

func NewAppWithDB(cfg *config.Config, database *gorm.DB) (*App, error) {
	redisClient, err := rdb.Connect(cfg.RedisConfig)
	if err != nil {
		slog.Error("Failed to connect to redis in monitor-service", "err", err)
		return nil, err
	}

	esClient, err := es.Connect(cfg.ESConfig)
	if err != nil {
		slog.Error("Failed to connect to elasticsearch in monitor-service", "err", err)
		return nil, err
	}

	// Kafka Writers
	pingWriter := pkgmq.NewWriter(cfg.PingRequestWriterConfig)
	mailWriter := pkgmq.NewWriter(cfg.MailWriterConfig)

	// Kafka Readers
	serverEventsReader := pkgmq.NewReader(cfg.ServerEventReaderConfig)
	heartbeatReader := pkgmq.NewReader(cfg.HeartbeatReaderConfig)
	pingResponseReader := pkgmq.NewReader(cfg.PingResponseReaderConfig)

	return &App{
		Config:             cfg,
		DB:                 database,
		RedisClient:        redisClient,
		ESClient:           esClient,
		ServerEventsReader: serverEventsReader,
		HeartbeatReader:    heartbeatReader,
		PingResponseReader: pingResponseReader,
		PingRequestWriter:  pingWriter,
		MailWriter:         mailWriter,
	}, nil
}
