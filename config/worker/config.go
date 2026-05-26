package workerconfig

import (
	"os"
	"strconv"

	commonconfig "github.com/LeHuuHai/server-management/config/common"
	"github.com/joho/godotenv"
)

type Config struct {
	AppConfig    *AppConfig
	KafkaConfig  *commonconfig.KafkaConfig
	SenderConfig *GomailConfig
}

type AppConfig struct {
	NumThread int
	ReportURL string
}

type GomailConfig struct {
	Addr     string
	Port     int
	From     string
	Password string
}

func Load() (*Config, error) {
	err := godotenv.Load("./config/worker/.env.worker")
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
			ReportURL: os.Getenv("APP_REPORT_URL"),
		},
		KafkaConfig: &commonconfig.KafkaConfig{
			Writer: &commonconfig.KafkaWriterConfig{
				Broker: os.Getenv("KAFKA_BROKER"),
			},
			Reader: &commonconfig.KafkaReaderConfig{
				Broker:          os.Getenv("KAFKA_BROKER"),
				ConsumerGroupId: os.Getenv("KAFKA_GROUP_ID"),
			},
			Topics: map[string]string{
				"ping":     os.Getenv("KAFKA_PING_TOPIC"),
				"mail":     os.Getenv("KAFKA_MAIL_TOPIC"),
				"ping_res": os.Getenv("KAFKA_HEARTBEAT_TOPIC"),
			},
		},
		SenderConfig: &GomailConfig{
			Addr:     os.Getenv("GOMAIL_ADDR"),
			Port:     gomailPort,
			From:     os.Getenv("GOMAIL_FROM"),
			Password: os.Getenv("GOMAIL_PASSWORD"),
		},
	}, nil

}
