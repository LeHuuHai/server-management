package kfk

import (
	"context"

	"github.com/LeHuuHai/server-management/internal/domain/mq"
	"github.com/segmentio/kafka-go"
)

type KfkPublisher struct {
	writer *kafka.Writer
}

func (p *KfkPublisher) Publish(ctx context.Context, msg mq.Message) error {
	err := p.writer.WriteMessages(
		ctx,
		kafka.Message{
			Topic: msg.Topic,
			Value: msg.Value,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func NewPublisher(w *kafka.Writer) *KfkPublisher {
	return &KfkPublisher{writer: w}
}
