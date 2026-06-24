package kafka

import (
	"context"
	"encoding/json"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/mq"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
	"github.com/segmentio/kafka-go"
)

type LifecycleConsumer struct {
	reader *kafka.Reader
}

func NewLifecycleConsumer(r *kafka.Reader) mq.LifecycleConsumerInterface {
	return &LifecycleConsumer{
		reader: r,
	}
}

func (c *LifecycleConsumer) Read(ctx context.Context) (pkgmodel.ServerEvent, string, error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return pkgmodel.ServerEvent{}, "", err
	}

	var event pkgmodel.ServerEvent
	err = json.Unmarshal(msg.Value, &event)
	if err != nil {
		return pkgmodel.ServerEvent{}, "", err
	}

	eventType := ""
	for _, h := range msg.Headers {
		if h.Key == "event_type" {
			eventType = string(h.Value)
			break
		}
	}

	action := "Update"
	if eventType == "ServerCreated" {
		action = "Create"
	} else if eventType == "ServerDeleted" {
		action = "Delete"
	}

	return event, action, nil
}

func (c *LifecycleConsumer) Close() error {
	return c.reader.Close()
}
