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

func (c *LifecycleConsumer) Read(ctx context.Context) (pkgmodel.ServerEvent, pkgmodel.ServerActionType, func(context.Context) error, error) {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return pkgmodel.ServerEvent{}, "", nil, err
	}

	var event pkgmodel.ServerEvent
	err = json.Unmarshal(msg.Value, &event)
	if err != nil {
		return pkgmodel.ServerEvent{}, "", nil, err
	}

	eventType := ""
	for _, h := range msg.Headers {
		if h.Key == "event_type" {
			eventType = string(h.Value)
			break
		}
	}

	action := pkgmodel.ServerUpdateAction
	switch pkgmodel.ServerEventType(eventType) {
	case pkgmodel.ServerCreateEvent:
		action = pkgmodel.ServerCreateAction
	case pkgmodel.ServerDeleteEvent:
		action = pkgmodel.ServerDeleteAction
	}

	return event, pkgmodel.ServerActionType(action), func(ctx context.Context) error {
		return c.reader.CommitMessages(ctx, msg)
	}, nil
}

func (c *LifecycleConsumer) Close() error {
	return c.reader.Close()
}
