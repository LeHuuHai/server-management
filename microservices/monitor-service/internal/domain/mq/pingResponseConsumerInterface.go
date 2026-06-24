package mq

import (
	"context"

	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type PingResponseConsumerInterface interface {
	Read(ctx context.Context) (pkgmodel.ResponsePing, error)
	Close() error
}
