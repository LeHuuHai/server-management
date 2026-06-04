package workerconfig

import (
	"log/slog"
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
	ReportKey string
}

type GomailConfig struct {
	Addr     string
	Port     int
	From     string
	Password string
}

func (c *AppConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("num_thread", c.NumThread),
		slog.String("report_url", c.ReportURL),
		slog.String("report_key", c.ReportKey),
	)
}

func (c *GomailConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("addr", c.Addr),
		slog.Int("port", c.Port),
		slog.String("from", c.From),
		slog.String("password", c.Password),
	)
}

func (c *Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("app", c.AppConfig),
		slog.Any("kafka", c.KafkaConfig),
		slog.Any("sender", c.SenderConfig),
	)
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

	cfg := Config{
		AppConfig: &AppConfig{
			NumThread: appNumThread,
			ReportURL: os.Getenv("APP_REPORT_URL"),
			ReportKey: os.Getenv("APP_REPORT_KEY"),
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
				"ping_res": os.Getenv("KAFKA_PING_RES_TOPIC"),
			},
		},
		SenderConfig: &GomailConfig{
			Addr:     os.Getenv("GOMAIL_ADDR"),
			Port:     gomailPort,
			From:     os.Getenv("GOMAIL_FROM"),
			Password: os.Getenv("GOMAIL_PASSWORD"),
		},
	}
	slog.Info("Config loaded", "config", &cfg)
	return &cfg, nil

}
