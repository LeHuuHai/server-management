package config

import (
	"log/slog"
	"os"
	"strconv"

	pkgconfig "github.com/LeHuuHai/server-management/microservices/pkg/config"
)

type AppConfig struct {
	Port             int
	GRPCPort         int
	Host             string
	CyclePing        int // in milliseconds
	HeartbeatTimeout int // in milliseconds
	AdMail           string
	InternalAPIKey   string
}

func (c *AppConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("port", c.Port),
		slog.Any("grpc_port", c.GRPCPort),
		slog.Any("host", c.Host),
		slog.Any("cycle ping", c.CyclePing),
		slog.Any("heartbeat timeout", c.HeartbeatTimeout),
		slog.Any("admin mail", c.AdMail),
		slog.Any("internal api key", c.InternalAPIKey),
	)
}

type Config struct {
	AppConfig                *AppConfig
	DBConfig                 *pkgconfig.PostgresConfig
	RedisConfig              *pkgconfig.RedisConfig
	ESConfig                 *pkgconfig.ElasticsearchConfig
	ServerEventReaderConfig  *pkgconfig.KafkaReaderConfig
	HeartbeatReaderConfig    *pkgconfig.KafkaReaderConfig
	PingResponseReaderConfig *pkgconfig.KafkaReaderConfig
	PingRequestWriterConfig  *pkgconfig.KafkaWriterConfig
	MailWriterConfig         *pkgconfig.KafkaWriterConfig
}

func (c *Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("app config", c.AppConfig),
		slog.Any("db config", c.DBConfig),
		slog.Any("redis config", c.RedisConfig),
		slog.Any("es config", c.ESConfig),
		slog.Any("server event reader config", c.ServerEventReaderConfig),
		slog.Any("heartbeat reader config", c.HeartbeatReaderConfig),
		slog.Any("ping response reader config", c.PingResponseReaderConfig),
		slog.Any("ping request writer config", c.PingRequestWriterConfig),
		slog.Any("mail writer config", c.MailWriterConfig),
	)
}

func Load() (*Config, error) {
	appPort, err := strconv.Atoi(os.Getenv("APP_PORT"))
	if err != nil {
		return nil, err
	}

	appGrpcPort, err := strconv.Atoi(os.Getenv("APP_GRPC_PORT"))
	if err != nil {
		return nil, err
	}

	cyclePing, err := strconv.Atoi(os.Getenv("CYCLE_PING"))
	if err != nil {
		return nil, err
	}

	hbTimeout, err := strconv.Atoi(os.Getenv("HEARTBEAT_TIMEOUT"))
	if err != nil {
		return nil, err
	}

	pgport, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return nil, err
	}

	redisDb, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		return nil, err
	}

	cfg := Config{
		AppConfig: &AppConfig{
			Port:             appPort,
			GRPCPort:         appGrpcPort,
			Host:             os.Getenv("APP_HOST"),
			CyclePing:        cyclePing,
			HeartbeatTimeout: hbTimeout,
			AdMail:           os.Getenv("AD_MAIL"),
			InternalAPIKey:   pkgconfig.ReadSecret("download_report_api_key"),
		},
		DBConfig: &pkgconfig.PostgresConfig{
			Host:     os.Getenv("DB_HOST"),
			Username: os.Getenv("DB_USER"),
			Password: pkgconfig.ReadSecret("monitor_db_password"),
			Port:     pgport,
			Database: os.Getenv("DB_DBNAME"),
		},
		RedisConfig: &pkgconfig.RedisConfig{
			URL: os.Getenv("REDIS_URL"),
			DB:  redisDb,
		},
		ESConfig: &pkgconfig.ElasticsearchConfig{
			URL:   os.Getenv("ES_URL"),
			Index: os.Getenv("ES_INDEX"),
		},
		ServerEventReaderConfig: &pkgconfig.KafkaReaderConfig{
			Broker:  os.Getenv("KAFKA_BROKER"),
			Topic:   os.Getenv("KAFKA_SERVER_TOPIC"),
			GroupID: os.Getenv("KAFKA_SERVER_EVENT_GROUP_ID"),
		},
		HeartbeatReaderConfig: &pkgconfig.KafkaReaderConfig{
			Broker:  os.Getenv("KAFKA_BROKER"),
			Topic:   os.Getenv("KAFKA_HEARTBEAT_TOPIC"),
			GroupID: os.Getenv("KAFKA_HEARTBEAT_GROUP_ID"),
		},
		PingResponseReaderConfig: &pkgconfig.KafkaReaderConfig{
			Broker:  os.Getenv("KAFKA_BROKER"),
			Topic:   os.Getenv("KAFKA_PING_RES_TOPIC"),
			GroupID: os.Getenv("KAFKA_PING_RES_GROUP_ID"),
		},
		PingRequestWriterConfig: &pkgconfig.KafkaWriterConfig{
			Broker: os.Getenv("KAFKA_BROKER"),
			Topic:  os.Getenv("KAFKA_PING_TOPIC"),
		},
		MailWriterConfig: &pkgconfig.KafkaWriterConfig{
			Broker: os.Getenv("KAFKA_BROKER"),
			Topic:  os.Getenv("KAFKA_MAIL_TOPIC"),
		},
	}

	return &cfg, nil
}
