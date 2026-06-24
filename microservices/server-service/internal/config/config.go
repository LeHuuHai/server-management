package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"

	pkgconfig "github.com/LeHuuHai/server-management/microservices/pkg/config"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	Port        int
	Host        string
	CORSOrigins []string
}

type Config struct {
	AppConfig   *AppConfig
	DBConfig    *pkgconfig.PostgresConfig
	KafkaConfig *pkgconfig.KafkaWriterConfig
}

func Load() (*Config, error) {
	err := godotenv.Load("./microservices/server-service/.env")
	if err != nil {
		slog.Warn("Error loading .env file, fallback to os env")
	}

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
			Password: os.Getenv("DB_PASSWORD"),
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
