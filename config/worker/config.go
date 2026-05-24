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
	KafkaReaderConfig *KafkaReaderConfig
	SenderConfig      *GomailConfig
}

type AppConfig struct {
	NumThread int
}

type KafkaReaderConfig struct {
	Broker          string
	PingTopic       string
	MailTopic       string
	ConsumerGroupId string
}

type GomailConfig struct {
	Addr     string
	Port     int
	From     string
	Password string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	appNumThread, err := strconv.Atoi(os.Getenv("APP_NUM_THREAD"))
	if err != nil {
		return nil, err
	}

	gomailPort, err := strconv.Atoi(os.Getenv("GOMAIL_PORT"))
	if err != nil {
		return nil, err
	}

	return &Config{
		AppConfig: &AppConfig{
			NumThread: appNumThread,
		},
		KafkaWriterConfig: &commonconfig.KafkaWriterConfig{
			Broker: os.Getenv("KAFKA_BROKER"),
		},
		KafkaReaderConfig: &KafkaReaderConfig{
			Broker:          os.Getenv("KAFKA_BROKER"),
			ConsumerGroupId: os.Getenv("KAFKA_GROUP_ID"),
			PingTopic:       os.Getenv("KAFKA_PING_TOPIC"),
			MailTopic:       os.Getenv("KAFKA_MAIL_TOPIC"),
		},
		SenderConfig: &GomailConfig{
			Addr:     os.Getenv("GOMAIL_ADDR"),
			Port:     gomailPort,
			From:     os.Getenv("GOMAIL_FROM"),
			Password: os.Getenv("GOMAIL_PASSWORD"),
		},
	}, nil

}
