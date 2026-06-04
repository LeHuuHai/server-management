package gwconfig

import (
	"log/slog"
	"os"
	"strconv"

	commonconfig "github.com/LeHuuHai/server-management/config/common"
	"github.com/joho/godotenv"
)

type Config struct {
	AppConfig   *AppConfig
	KafkaConfig *commonconfig.KafkaConfig
}

type AppConfig struct {
	Port         int
	Host         string
	HeartbeatKey string
}

func (c *AppConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("port", c.Port),
		slog.String("host", c.Host),
		slog.String("heartbeat_key", c.HeartbeatKey),
	)
}

func (c *Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("app", c.AppConfig),
		slog.Any("kafka", c.KafkaConfig),
	)
}

func Load() (*Config, error) {
	err := godotenv.Load("./config/gw/.env.gw")
	if err != nil {
		panic("Error loading .env file")
	}

	appPort, err := strconv.Atoi(os.Getenv("APP_PORT"))
	if err != nil {
		return nil, err
	}

	cfg := Config{
		AppConfig: &AppConfig{
			Port:         appPort,
			Host:         os.Getenv("APP_HOST"),
			HeartbeatKey: os.Getenv("APP_HEARTBEAT_KEY"),
		},
		KafkaConfig: &commonconfig.KafkaConfig{
			Writer: &commonconfig.KafkaWriterConfig{
				Broker: os.Getenv("KAFKA_BROKER"),
			},
			Topics: map[string]string{
				"heartbeat": os.Getenv("KAFKA_HEARTBEAT_TOPIC"),
			},
		},
	}
	slog.Info("Config loaded", "config", &cfg)
	return &cfg, nil
}
