package mq

import (
	"context"

	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type PublisherInterface interface {
	Publish(ctx context.Context, hb pkgmodel.Heartbeat) error
}
