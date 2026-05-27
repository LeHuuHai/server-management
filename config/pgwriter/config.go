package pgwriterconfig

import (
	"os"
	"strconv"

	commonconfig "github.com/LeHuuHai/server-management/config/common"
	"github.com/joho/godotenv"
)

type Config struct {
	AppConfig   *AppConfig
	KafkaConfig *commonconfig.KafkaConfig
	DBConfig    *commonconfig.PostgresConfig
}

type AppConfig struct {
}

func Load() (*Config, error) {
	err := godotenv.Load("./config/pgwriter/.env.pgwriter")
	if err != nil {
		panic("Error loading .env file")
	}

	pgport, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return nil, err
	}

	return &Config{
		AppConfig: &AppConfig{},
		KafkaConfig: &commonconfig.KafkaConfig{
			Reader: &commonconfig.KafkaReaderConfig{
				Broker:          os.Getenv("KAFKA_BROKER"),
				ConsumerGroupId: os.Getenv("KAFKA_GROUP_ID"),
			},
			Topics: map[string]string{
				"ping_res": os.Getenv("KAFKA_HEARTBEAT_TOPIC"),
			},
		},
		DBConfig: &commonconfig.PostgresConfig{
			Host:     os.Getenv("DB_HOST"),
			Username: os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Port:     pgport,
			Database: os.Getenv("DB_DBNAME"),
		},
	}, nil
}
