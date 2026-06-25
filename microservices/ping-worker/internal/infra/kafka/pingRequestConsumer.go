package kafka

import (
	"context"
	"encoding/json"

	"github.com/LeHuuHai/server-management/microservices/ping-worker/internal/domain/mq"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
	"github.com/segmentio/kafka-go"
)

type PingRequestConsumer struct {
	reader *kafka.Reader
}

func NewPingRequestConsumer(r *kafka.Reader) mq.PingRequestConsumerInterface {
	return &PingRequestConsumer{
		reader: r,
	}
}

func (c *PingRequestConsumer) Read(ctx context.Context) (pkgmodel.RequestPing, func(context.Context) error, error) {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return pkgmodel.RequestPing{}, nil, err
	}

	var req pkgmodel.RequestPing
	err = json.Unmarshal(msg.Value, &req)
	if err != nil {
		return pkgmodel.RequestPing{}, nil, err
	}

	commitFunc := func(ctx context.Context) error {
		return c.reader.CommitMessages(ctx, msg)
	}

	return req, commitFunc, nil
}

func (c *PingRequestConsumer) Close() error {
	return c.reader.Close()
}
