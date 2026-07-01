package config

import (
	"log/slog"
	"os"
	"strconv"

	pkgconfig "github.com/LeHuuHai/server-management/microservices/pkg/config"
)

type AppConfig struct {
	NumThread int
}

func (c *AppConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("num thread", c.NumThread),
	)
}

type Config struct {
	AppConfig                *AppConfig
	PingRequestReaderConfig  *pkgconfig.KafkaReaderConfig
	PingResponseWriterConfig *pkgconfig.KafkaWriterConfig
}

func (c *Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("app config", c.AppConfig),
		slog.Any("ping request reader config", c.PingRequestReaderConfig),
		slog.Any("ping response writer config", c.PingResponseWriterConfig),
	)
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
			Broker:  os.Getenv("KAFKA_BROKER"),
			Topic:   os.Getenv("KAFKA_PING_TOPIC"),
			GroupID: os.Getenv("KAFKA_PING_REQ_GROUP_ID"),
		},
		PingResponseWriterConfig: &pkgconfig.KafkaWriterConfig{
			Broker: os.Getenv("KAFKA_BROKER"),
			Topic:  os.Getenv("KAFKA_PING_RES_TOPIC"),
		},
	}

	return &cfg, nil
}
