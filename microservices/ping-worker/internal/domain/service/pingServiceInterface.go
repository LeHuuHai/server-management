package service

import (
	"context"

	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type PingServiceInterface interface {
	Ping(ctx context.Context, req pkgmodel.RequestPing) (pkgmodel.ResponsePing, error)
}
