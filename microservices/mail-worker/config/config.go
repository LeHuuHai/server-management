package config

import (
	"log/slog"
	"os"
	"strconv"

	pkgconfig "github.com/LeHuuHai/server-management/microservices/pkg/config"
)

type SenderConfig struct {
	Addr     string
	Port     int
	From     string
	Password string
}

func (c *SenderConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("addr", c.Addr),
		slog.Any("port", c.Port),
		slog.Any("from", c.From),
		slog.Any("password", c.Password),
	)
}

type AppConfig struct {
	ReportRepoAddr string
	InternalAPIKey string
}

func (c *AppConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("report repo addr", c.ReportRepoAddr),
		slog.Any("internal api key", c.InternalAPIKey),
	)
}

type Config struct {
	AppConfig        *AppConfig
	MailReaderConfig *pkgconfig.KafkaReaderConfig
	SenderConfig     *SenderConfig
}

func (c *Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("app config", c.AppConfig),
		slog.Any("mail reader config", c.MailReaderConfig),
		slog.Any("sender config", c.SenderConfig),
	)
}

func Load() (*Config, error) {
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
