package kfk

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type kafkaPublisher struct {
	writer *kafka.Writer
}

func (p *kafkaPublisher) Publish(ctx context.Context, msg []byte) error {
	err := p.writer.WriteMessages(
		ctx,
		kafka.Message{
			Value: msg,
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
