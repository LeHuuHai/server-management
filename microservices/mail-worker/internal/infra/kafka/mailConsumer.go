package kafka

import (
	"context"
	"encoding/json"

	"github.com/LeHuuHai/server-management/microservices/mail-worker/internal/domain/mq"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
	"github.com/segmentio/kafka-go"
)

type MailConsumer struct {
	reader *kafka.Reader
}

func NewMailConsumer(r *kafka.Reader) mq.MailConsumerInterface {
	return &MailConsumer{
		reader: r,
	}
}

func (c *MailConsumer) Read(ctx context.Context) (pkgmodel.RequestMail, func(context.Context) error, error) {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return pkgmodel.RequestMail{}, nil, err
	}

	var req pkgmodel.RequestMail
	err = json.Unmarshal(msg.Value, &req)
	if err != nil {
		return pkgmodel.RequestMail{}, nil, err
	}

	commitFunc := func(ctx context.Context) error {
		return c.reader.CommitMessages(ctx, msg)
	}

	return req, commitFunc, nil
}

func (c *MailConsumer) Close() error {
	return c.reader.Close()
}
