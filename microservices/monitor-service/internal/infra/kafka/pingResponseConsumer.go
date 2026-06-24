package kafka

import (
	"context"
	"encoding/json"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/mq"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
	"github.com/segmentio/kafka-go"
)

type PingResponseConsumer struct {
	reader *kafka.Reader
}

func NewPingResponseConsumer(r *kafka.Reader) mq.PingResponseConsumerInterface {
	return &PingResponseConsumer{
		reader: r,
	}
}

func (c *PingResponseConsumer) Read(ctx context.Context) (pkgmodel.ResponsePing, error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return pkgmodel.ResponsePing{}, err
	}

	var res pkgmodel.ResponsePing
	err = json.Unmarshal(msg.Value, &res)
	if err != nil {
		return pkgmodel.ResponsePing{}, err
	}

	return res, nil
}

func (c *PingResponseConsumer) Close() error {
	return c.reader.Close()
}
