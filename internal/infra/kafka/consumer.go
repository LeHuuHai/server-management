package kfk

import (
	"context"
	"fmt"

	"github.com/LeHuuHai/server-management/internal/domain/mq"
	"github.com/segmentio/kafka-go"
)

type KfkConsumer struct {
	reader *kafka.Reader
}

func (c *KfkConsumer) Read(ctx context.Context) (*mq.Message, error) {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return nil, err
	}
	return &mq.Message{
		Topic: msg.Topic,
		Value: msg.Value,
		Raw:   &msg,
	}, nil
}

func (c *KfkConsumer) Commit(ctx context.Context, msg *mq.Message) error {
	if msg.Raw == nil {
		return fmt.Errorf("nil raw kafka message")
	}
	err := c.reader.CommitMessages(ctx, *msg.Raw)
	if err != nil {
		panic(err)
	}
	return nil
}

func NewConsumer(r *kafka.Reader) *KfkConsumer {
	return &KfkConsumer{
		reader: r,
	}
}
