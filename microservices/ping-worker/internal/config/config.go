package config

import (
	"log/slog"
	"os"
	"strconv"

	pkgconfig "github.com/LeHuuHai/server-management/microservices/pkg/config"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	NumThread int
}

type Config struct {
	AppConfig                *AppConfig
	PingRequestReaderConfig  *pkgconfig.KafkaReaderConfig
	PingResponseWriterConfig *pkgconfig.KafkaWriterConfig
}

func Load() (*Config, error) {
	err := godotenv.Load("./microservices/ping-worker/.env")
	if err != nil {
		slog.Warn("Error loading .env file, fallback to os env")
	}

	numThread, err := strconv.Atoi(os.Getenv("APP_NUM_THREAD"))
	if err != nil {
		return nil, err
	}

	cfg := Config{
		AppConfig: &AppConfig{
			NumThread: numThread,
		},
		PingRequestReaderConfig: &pkgconfig.KafkaReaderConfig{
			Broker:     os.Getenv("KAFKA_BROKER"),
			Topic:      os.Getenv("KAFKA_PING_TOPIC"),
			ConsumerID: os.Getenv("KAFKA_CONSUMER_GROUP"),
		},
		PingResponseWriterConfig: &pkgconfig.KafkaWriterConfig{
			Broker: os.Getenv("KAFKA_BROKER"),
			Topic:  os.Getenv("KAFKA_PING_RES_TOPIC"),
		},
	}

	return &cfg, nil
}
