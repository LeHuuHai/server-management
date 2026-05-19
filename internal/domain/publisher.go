package domain

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Publisher struct {
	writer *kafka.Writer
}

func (p *Publisher) Publish(ctx context.Context, msg []byte) error {
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

func NewPublisher(w *kafka.Writer) *Publisher {
	return &Publisher{writer: w}
}
