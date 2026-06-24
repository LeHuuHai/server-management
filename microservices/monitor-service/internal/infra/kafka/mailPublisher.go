package kafka

import (
	"context"
	"encoding/json"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/mq"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
	"github.com/segmentio/kafka-go"
)

type MailPublisher struct {
	writer *kafka.Writer
}

func NewMailPublisher(w *kafka.Writer) mq.MailPublisherInterface {
	return &MailPublisher{
		writer: w,
	}
}

func (p *MailPublisher) Publish(ctx context.Context, req pkgmodel.RequestMail) error {
	bytes, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Value: bytes,
	})
}
