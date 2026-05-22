package kfk

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type kafkaPublisher struct {
	writer *kafka.Writer
}

func (p *kafkaPublisher) Publish(ctx context.Context, topic string, msg []byte) error {
	err := p.writer.WriteMessages(
		ctx,
		kafka.Message{
			Topic: topic,
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
