package kafka

import (
	"context"
	"encoding/json"

	"github.com/LeHuuHai/server-management/microservices/ping-worker/internal/domain/mq"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
	"github.com/segmentio/kafka-go"
)

type PingResponsePublisher struct {
	writer *kafka.Writer
}

func NewPingResponsePublisher(w *kafka.Writer) mq.PingResponsePublisherInterface {
	return &PingResponsePublisher{
		writer: w,
	}
}

func (p *PingResponsePublisher) Publish(ctx context.Context, response pkgmodel.ResponsePing) error {
	bytes, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(response.ServerID),
		Value: bytes,
	})
}
