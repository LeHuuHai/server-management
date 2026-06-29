package config

import (
	"log/slog"
	"os"
	"strconv"

	pkgconfig "github.com/LeHuuHai/server-management/microservices/pkg/config"
	"github.com/joho/godotenv"
)

type SenderConfig struct {
	Addr     string
	Port     int
	From     string
	Password string
}

type AppConfig struct {
	ReportRepoAddr string
	InternalAPIKey string
}

type Config struct {
	AppConfig        *AppConfig
	MailReaderConfig *pkgconfig.KafkaReaderConfig
	SenderConfig     *SenderConfig
}

func Load() (*Config, error) {
	err := godotenv.Load("./microservices/mail-worker/.env")
	if err != nil {
		slog.Warn("Error loading .env file, fallback to os env")
	}

	gomailPort, err := strconv.Atoi(os.Getenv("GOMAIL_PORT"))
	if err != nil {
		return nil, err
	}

	cfg := Config{
		AppConfig: &AppConfig{
			ReportRepoAddr: os.Getenv("REPORT_REPO_ADDR"),
			InternalAPIKey: pkgconfig.ReadSecret("download_report_api_key"),
		},
		MailReaderConfig: &pkgconfig.KafkaReaderConfig{
			Broker:     os.Getenv("KAFKA_BROKER"),
			Topic:      os.Getenv("KAFKA_MAIL_TOPIC"),
			ConsumerID: os.Getenv("KAFKA_CONSUMER_GROUP"),
		},
		SenderConfig: &SenderConfig{
			Addr:     os.Getenv("GOMAIL_ADDR"),
			Port:     gomailPort,
			From:     os.Getenv("GOMAIL_FROM"),
			Password: pkgconfig.ReadSecret("gomail_password"),
		},
	}

	return &cfg, nil
}
