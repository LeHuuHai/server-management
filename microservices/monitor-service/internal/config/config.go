package config

import (
	"log/slog"
	"os"
	"strconv"

	pkgconfig "github.com/LeHuuHai/server-management/microservices/pkg/config"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	Port             int
	Host             string
	CyclePing        int // in milliseconds
	HeartbeatTimeout int // in milliseconds
	AdMail           string
	InternalAPIKey   string
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

func Load() (*Config, error) {
	err := godotenv.Load("./microservices/monitor-service/.env")
	if err != nil {
		slog.Warn("Error loading .env file, fallback to os env")
	}

	appPort, err := strconv.Atoi(os.Getenv("APP_PORT"))
	if err != nil {
		appPort = 50054
	}

	cyclePing, err := strconv.Atoi(os.Getenv("CYCLE_PING"))
	if err != nil {
		cyclePing = 5000
	}

	hbTimeout, err := strconv.Atoi(os.Getenv("HEARTBEAT_TIMEOUT"))
	if err != nil {
		hbTimeout = 10000
	}

	pgport, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		pgport = 5432
	}

	redisDb, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		redisDb = 0
	}

	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}

	group := os.Getenv("KAFKA_CONSUMER_GROUP")
	if group == "" {
		group = "monitor-service-group"
	}

	serverTopic := os.Getenv("KAFKA_SERVER_TOPIC")
	if serverTopic == "" {
		serverTopic = "server_events"
	}

	hbTopic := os.Getenv("KAFKA_HEARTBEAT_TOPIC")
	if hbTopic == "" {
		hbTopic = "heartbeat"
	}

	pingTopic := os.Getenv("KAFKA_PING_TOPIC")
	if pingTopic == "" {
		pingTopic = "ping"
	}

	pingResTopic := os.Getenv("KAFKA_PING_RES_TOPIC")
	if pingResTopic == "" {
		pingResTopic = "ping_res"
	}

	mailTopic := os.Getenv("KAFKA_MAIL_TOPIC")
	if mailTopic == "" {
		mailTopic = "mail"
	}

	internalAPIKey := pkgconfig.ReadSecret("internal_api_key")
	if internalAPIKey == "" {
		internalAPIKey = "internal-secret-key"
	}

	cfg := Config{
		AppConfig: &AppConfig{
			Port:             appPort,
			Host:             os.Getenv("APP_HOST"),
			CyclePing:        cyclePing,
			HeartbeatTimeout: hbTimeout,
			AdMail:           os.Getenv("AD_MAIL"),
			InternalAPIKey:   internalAPIKey,
		},
		DBConfig: &pkgconfig.PostgresConfig{
			Host:     os.Getenv("DB_HOST"),
			Username: os.Getenv("DB_USER"),
			Password: pkgconfig.ReadSecret("monitor_db_password"),
			Port:     pgport,
			Database: os.Getenv("DB_DBNAME"),
		},
		RedisConfig: &pkgconfig.RedisConfig{
			URL:      os.Getenv("REDIS_URL"),
			Password: pkgconfig.ReadSecret("redis_password"),
			DB:       redisDb,
		},
		ESConfig: &pkgconfig.ElasticsearchConfig{
			URL:      os.Getenv("ES_URL"),
			Username: os.Getenv("ES_USER"),
			Password: pkgconfig.ReadSecret("es_password"),
			Index:    os.Getenv("ES_INDEX"),
		},
		ServerEventReaderConfig: &pkgconfig.KafkaReaderConfig{
			Broker:     broker,
			Topic:      serverTopic,
			ConsumerID: group + "-lifecycle",
		},
		HeartbeatReaderConfig: &pkgconfig.KafkaReaderConfig{
			Broker:     broker,
			Topic:      hbTopic,
			ConsumerID: group + "-heartbeat",
		},
		PingResponseReaderConfig: &pkgconfig.KafkaReaderConfig{
			Broker:     broker,
			Topic:      pingResTopic,
			ConsumerID: group + "-ping-res",
		},
		PingRequestWriterConfig: &pkgconfig.KafkaWriterConfig{
			Broker: broker,
			Topic:  pingTopic,
		},
		MailWriterConfig: &pkgconfig.KafkaWriterConfig{
			Broker: broker,
			Topic:  mailTopic,
		},
	}

	return &cfg, nil
}
