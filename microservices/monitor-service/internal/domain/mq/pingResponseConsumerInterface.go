package mq

import (
	"context"

	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type PingResponseConsumerInterface interface {
	Read(ctx context.Context) (pkgmodel.ResponsePing, func(context.Context) error, error)
	Close() error
}
