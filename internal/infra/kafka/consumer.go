package kfk

import (
	"context"

	"github.com/LeHuuHai/server-management/internal/domain/mq"
	"github.com/segmentio/kafka-go"
)

type KfkConsumer struct {
	reader *kafka.Reader
}

func (c *KfkConsumer) Read(ctx context.Context) (*mq.Message, error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}
	return &mq.Message{
		Topic: msg.Topic,
		Value: msg.Value,
	}, nil
}

func NewConsumer(r *kafka.Reader) *KfkConsumer {
	return &KfkConsumer{
		reader: r,
	}
}
