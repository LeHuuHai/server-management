package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"

	pkgconfig "github.com/LeHuuHai/server-management/microservices/pkg/config"
)

type AppConfig struct {
	Port        int
	Host        string
	CORSOrigins []string
}

func (c *AppConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("host", c.Host),
		slog.Any("port", c.Port),
	)
}

type Config struct {
	AppConfig   *AppConfig
	DBConfig    *pkgconfig.PostgresConfig
	KafkaConfig *pkgconfig.KafkaWriterConfig
}

func (c *Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("app config", c.AppConfig),
		slog.Any("db config", c.DBConfig),
		slog.Any("kafka config", c.KafkaConfig),
	)
}

func Load() (*Config, error) {
	pgport, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		pgport = 5432
	}

	appPort, err := strconv.Atoi(os.Getenv("APP_PORT"))
	if err != nil {
		appPort = 50052
	}

	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}

	topic := os.Getenv("KAFKA_SERVER_TOPIC")
	if topic == "" {
		topic = "server_events"
	}

	cfg := Config{
		AppConfig: &AppConfig{
			Port:        appPort,
			Host:        os.Getenv("APP_HOST"),
			CORSOrigins: strings.Split(os.Getenv("APP_CORS_ORIGIN"), ","),
		},
		DBConfig: &pkgconfig.PostgresConfig{
			Host:     os.Getenv("DB_HOST"),
			Username: os.Getenv("DB_USER"),
			Password: pkgconfig.ReadSecret("server_db_password"),
			Port:     pgport,
			Database: os.Getenv("DB_DBNAME"),
		},
		KafkaConfig: &pkgconfig.KafkaWriterConfig{
			Broker: broker,
			Topic:  topic,
		},
	}
	return &cfg, nil
}
