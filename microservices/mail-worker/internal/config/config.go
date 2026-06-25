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
	MonitorRPC       string
}

func Load() (*Config, error) {
	err := godotenv.Load("./microservices/mail-worker/.env")
	if err != nil {
		slog.Warn("Error loading .env file, fallback to os env")
	}

	gomailPort, err := strconv.Atoi(os.Getenv("GOMAIL_PORT"))
	if err != nil {
		gomailPort = 587
	}

	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}

	group := os.Getenv("KAFKA_CONSUMER_GROUP")
	if group == "" {
		group = "mail-worker-group"
	}

	mailTopic := os.Getenv("KAFKA_MAIL_TOPIC")
	if mailTopic == "" {
		mailTopic = "mail"
	}

	monitorRPC := os.Getenv("MONITOR_RPC")
	if monitorRPC == "" {
		monitorRPC = "localhost:50054"
	}

	internalAPIKey := os.Getenv("INTERNAL_API_KEY")
	if internalAPIKey == "" {
		internalAPIKey = "internal-secret-key"
	}

	cfg := Config{
		AppConfig: &AppConfig{
			ReportRepoAddr: os.Getenv("REPORT_REPO_ADDR"),
			InternalAPIKey: internalAPIKey,
		},
		MailReaderConfig: &pkgconfig.KafkaReaderConfig{
			Broker:     broker,
			Topic:      mailTopic,
			ConsumerID: group,
		},
		SenderConfig: &SenderConfig{
			Addr:     os.Getenv("GOMAIL_ADDR"),
			Port:     gomailPort,
			From:     os.Getenv("GOMAIL_FROM"),
			Password: os.Getenv("GOMAIL_PASSWORD"),
		},
		MonitorRPC: monitorRPC,
	}

	return &cfg, nil
}
