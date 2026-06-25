package mq

import (
	"context"

	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type HeartbeatConsumerInterface interface {
	Read(ctx context.Context) (pkgmodel.Heartbeat, func(context.Context) error, error)
	Close() error
}
