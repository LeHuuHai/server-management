package kfk

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type publisher struct {
	writer *kafka.Writer
}

func (p *publisher) Publish(ctx context.Context, msg []byte) error {
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

func NewPublisher(w *kafka.Writer) *publisher {
	return &publisher{writer: w}
}
