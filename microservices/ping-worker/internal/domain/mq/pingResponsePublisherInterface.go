package mq

import (
	"context"

	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type PingResponsePublisherInterface interface {
	Publish(ctx context.Context, response pkgmodel.ResponsePing) error
}
