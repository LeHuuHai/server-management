package kafka

import (
	"context"
	"encoding/json"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/mq"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
	"github.com/segmentio/kafka-go"
)

type PingRequestPublisher struct {
	writer *kafka.Writer
}

func NewPingRequestPublisher(w *kafka.Writer) mq.PingRequestPublisherInterface {
	return &PingRequestPublisher{
		writer: w,
	}
}

func (p *PingRequestPublisher) Publish(ctx context.Context, req pkgmodel.RequestPing) error {
	bytes, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(req.ServerID),
		Value: bytes,
	})
}
