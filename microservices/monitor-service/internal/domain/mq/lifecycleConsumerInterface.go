package mq

import (
	"context"

	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type LifecycleConsumerInterface interface {
	Read(ctx context.Context) (event pkgmodel.ServerEvent, action string, err error)
	Close() error
}
