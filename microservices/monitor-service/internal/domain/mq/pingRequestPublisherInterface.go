package mq

import (
	"context"

	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type PingRequestPublisherInterface interface {
	Publish(ctx context.Context, req pkgmodel.RequestPing) error
}
