package rt

import (
	"fmt"

	"github.com/LeHuuHai/server-management/microservices/mail-worker/config"
	"github.com/LeHuuHai/server-management/microservices/pkg/apperr"
	"github.com/LeHuuHai/server-management/microservices/pkg/mq"
	"github.com/segmentio/kafka-go"
	"gopkg.in/gomail.v2"
)

type App struct {
	Config       *config.Config
	MailReader   *kafka.Reader
	GomailDialer *gomail.Dialer
}

func NewApp() (*App, error) {
	// load config
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrAppBuild, err)
	}

	// Initialize Kafka Reader
	mailReader := mq.NewReader(cfg.MailReaderConfig)

	// Initialize Gomail Dialer
	gomailDialer := gomail.NewDialer(
		cfg.SenderConfig.Addr,
		cfg.SenderConfig.Port,
		cfg.SenderConfig.From,
		cfg.SenderConfig.Password,
	)

	return &App{
		Config:       cfg,
		MailReader:   mailReader,
		GomailDialer: gomailDialer,
	}, nil
}
