package kafka

import (
	"context"
	"encoding/json"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/mq"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
	"github.com/segmentio/kafka-go"
)

type HeartbeatConsumer struct {
	reader *kafka.Reader
}

func NewHeartbeatConsumer(r *kafka.Reader) mq.HeartbeatConsumerInterface {
	return &HeartbeatConsumer{
		reader: r,
	}
}

func (c *HeartbeatConsumer) Read(ctx context.Context) (pkgmodel.Heartbeat, error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return pkgmodel.Heartbeat{}, err
	}

	var hb pkgmodel.Heartbeat
	err = json.Unmarshal(msg.Value, &hb)
	if err != nil {
		return pkgmodel.Heartbeat{}, err
	}

	return hb, nil
}

func (c *HeartbeatConsumer) Close() error {
	return c.reader.Close()
}
