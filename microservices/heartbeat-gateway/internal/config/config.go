package config

import (
	"log/slog"
	"os"
	"strconv"

	pkgconfig "github.com/LeHuuHai/server-management/microservices/pkg/config"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	Port int
	Host string
}

type Config struct {
	AppConfig   *AppConfig
	KafkaConfig *pkgconfig.KafkaConfig
}

func Load() (*Config, error) {
	err := godotenv.Load("./microservices/heartbeat-gateway/.env")
	if err != nil {
		slog.Warn("Error loading .env file, fallback to os env")
	}

	appPort, err := strconv.Atoi(os.Getenv("APP_PORT"))
	if err != nil {
		appPort = 50053
	}

	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}

	topic := os.Getenv("KAFKA_HEARTBEAT_TOPIC")
	if topic == "" {
		topic = "heartbeat"
	}

	cfg := Config{
		AppConfig: &AppConfig{
			Port: appPort,
			Host: os.Getenv("APP_HOST"),
		},
		KafkaConfig: &pkgconfig.KafkaConfig{
			Broker:      broker,
			ServerTopic: topic, // Using ServerTopic field from package config as the generic topic holder
		},
	}
	return &cfg, nil
}
