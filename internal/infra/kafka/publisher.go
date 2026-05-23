package kfk

import (
	"context"

	"github.com/LeHuuHai/server-management/internal/domain/mq"
	"github.com/segmentio/kafka-go"
)

type kafkaPublisher struct {
	writer *kafka.Writer
}

func (p *kafkaPublisher) Publish(ctx context.Context, msg mq.Message) error {
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

func NewPublisher(w *kafka.Writer) *kafkaPublisher {
	return &kafkaPublisher{writer: w}
}
