package kafka

import (
	"context"
	"encoding/json"

	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/internal/domain/mq"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPublisher(w *kafka.Writer) mq.PublisherInterface {
	return &KafkaPublisher{
		writer: w,
	}
}

func (p *KafkaPublisher) Publish(ctx context.Context, hb pkgmodel.Heartbeat) error {
	bytes, err := json.Marshal(hb)
	if err != nil {
		return err
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(hb.ServerID),
		Value: bytes,
	})
	return err
}
