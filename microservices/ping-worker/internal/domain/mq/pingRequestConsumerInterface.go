package mq

import (
	"context"

	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type PingRequestConsumerInterface interface {
	Read(ctx context.Context) (pkgmodel.RequestPing, func(context.Context) error, error)
	Close() error
}
