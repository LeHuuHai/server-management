package gwconfig

import (
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

func Load() (*Config, error) {
	err := godotenv.Load("./config/gw/.env.gw")
	if err != nil {
		panic("Error loading .env file")
	}

	appPort, err := strconv.Atoi(os.Getenv("APP_PORT"))
	if err != nil {
		return nil, err
	}

	return &Config{
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
	}, nil
}
