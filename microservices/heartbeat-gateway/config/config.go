package config

import (
	"log/slog"
	"os"
	"strconv"

	pkgconfig "github.com/LeHuuHai/server-management/microservices/pkg/config"
)

type AppConfig struct {
	Port         int
	Host         string
	HeartbeatKey string
}

func (c *AppConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("host", c.Host),
		slog.Any("port", c.Port),
		slog.Any("heartbeat key", c.HeartbeatKey),
	)
}

type Config struct {
	AppConfig   *AppConfig
	KafkaConfig *pkgconfig.KafkaWriterConfig
}

func (c *Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("app config", c.AppConfig),
		slog.Any("kafka config", c.KafkaConfig),
	)
}

func Load() (*Config, error) {
	appPort, err := strconv.Atoi(os.Getenv("APP_PORT"))
	if err != nil {
		return nil, err
	}
	cfg := Config{
		AppConfig: &AppConfig{
			Port:         appPort,
			Host:         os.Getenv("APP_HOST"),
			HeartbeatKey: pkgconfig.ReadSecret("heartbeat_api_key"),
		},
		KafkaConfig: &pkgconfig.KafkaWriterConfig{
			Broker: os.Getenv("KAFKA_BROKER"),
			Topic:  os.Getenv("KAFKA_HEARTBEAT_TOPIC"),
		},
	}
	return &cfg, nil
}
