package config

import (
	"os"
	"strconv"

	pkgconfig "github.com/LeHuuHai/server-management/microservices/pkg/config"
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
