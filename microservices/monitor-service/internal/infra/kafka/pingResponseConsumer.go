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

func (c *PingResponseConsumer) Read(ctx context.Context) (pkgmodel.ResponsePing, func(context.Context) error, error) {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return pkgmodel.ResponsePing{}, nil, err
	}

	var res pkgmodel.ResponsePing
	err = json.Unmarshal(msg.Value, &res)
	if err != nil {
		return pkgmodel.ResponsePing{}, nil, err
	}

	return res, func(ctx context.Context) error {
		return c.reader.CommitMessages(ctx, msg)
	}, nil
}

func (c *PingResponseConsumer) Close() error {
	return c.reader.Close()
}
