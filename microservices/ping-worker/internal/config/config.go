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
	AppConfig                 *AppConfig
	PingRequestReaderConfig   *pkgconfig.KafkaReaderConfig
	PingResponseWriterConfig  *pkgconfig.KafkaWriterConfig
}

func Load() (*Config, error) {
	err := godotenv.Load("./microservices/ping-worker/.env")
	if err != nil {
		slog.Warn("Error loading .env file, fallback to os env")
	}

	numThread, err := strconv.Atoi(os.Getenv("APP_NUM_THREAD"))
	if err != nil {
		numThread = 10
	}

	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}

	group := os.Getenv("KAFKA_CONSUMER_GROUP")
	if group == "" {
		group = "ping-worker-group"
	}

	pingTopic := os.Getenv("KAFKA_PING_TOPIC")
	if pingTopic == "" {
		pingTopic = "ping"
	}

	pingResTopic := os.Getenv("KAFKA_PING_RES_TOPIC")
	if pingResTopic == "" {
		pingResTopic = "ping_res"
	}

	cfg := Config{
		AppConfig: &AppConfig{
			NumThread: numThread,
		},
		PingRequestReaderConfig: &pkgconfig.KafkaReaderConfig{
			Broker:     broker,
			Topic:      pingTopic,
			ConsumerID: group,
		},
		PingResponseWriterConfig: &pkgconfig.KafkaWriterConfig{
			Broker: broker,
			Topic:  pingResTopic,
		},
	}

	return &cfg, nil
}
