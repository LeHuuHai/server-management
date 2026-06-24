package service

import (
	"context"

	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type GwServiceInterface interface {
	PublishHeartbeat(ctx context.Context, heartbeat pkgmodel.Heartbeat) error
}
