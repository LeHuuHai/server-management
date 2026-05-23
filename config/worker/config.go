package workerconfig

import (
	"os"
	"strconv"

	commonconfig "github.com/LeHuuHai/server-management/config/common"
	"github.com/joho/godotenv"
)

type Config struct {
	AppConfig         *AppConfig
	KafkaWriterConfig *commonconfig.KafkaWriterConfig
	KafkaReaderConfig *commonconfig.KafkaReaderConfig
	SenderConfig      *commonconfig.GomailConfig
}

type AppConfig struct {
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	gomailPort, err := strconv.Atoi(os.Getenv("GOMAIL_PORT"))
	if err != nil {
		return nil, err
	}

	return &Config{
		AppConfig: &AppConfig{},
		KafkaWriterConfig: &commonconfig.KafkaWriterConfig{
			Broker: os.Getenv("KAFKA_BROKER"),
		},
		KafkaReaderConfig: &commonconfig.KafkaReaderConfig{
			Broker:          os.Getenv("KAFKA_BROKER"),
			ConsumerGroupId: os.Getenv("KAFKA_GROUP_ID"),
			Topic:           os.Getenv("KAFKA_TOPIC"),
		},
		SenderConfig: &commonconfig.GomailConfig{
			Addr:     os.Getenv("GOMAIL_ADDR"),
			Port:     gomailPort,
			From:     os.Getenv("GOMAIL_FROM"),
			Password: os.Getenv("GOMAIL_PASSWORD"),
		},
	}, nil

}
