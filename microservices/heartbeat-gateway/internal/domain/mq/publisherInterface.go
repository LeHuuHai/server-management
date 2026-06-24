package mq

import (
	"context"

	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/internal/model"
)

type PublisherInterface interface {
	Publish(ctx context.Context, hb model.Heartbeat) error
}
