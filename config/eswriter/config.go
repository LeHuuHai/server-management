package eswriterconfig

import (
	"os"

	commonconfig "github.com/LeHuuHai/server-management/config/common"
	"github.com/joho/godotenv"
)

type Config struct {
	AppConfig   *AppConfig
	KafkaConfig *commonconfig.KafkaConfig
	ESConfig    *commonconfig.ElasticsearchConfig
}

type AppConfig struct {
}

func Load() (*Config, error) {
	err := godotenv.Load("./config/eswriter/.env.eswriter")
	if err != nil {
		panic("Error loading .env file")
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
		ESConfig: &commonconfig.ElasticsearchConfig{
			URL:   os.Getenv("ES_URL"),
			Index: os.Getenv("ES_INDEX"),
		},
	}, nil
}
