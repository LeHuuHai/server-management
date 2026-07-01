package service

import (
	"context"

	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type MonitorServiceInterface interface {
	ProcessHeartbeat(ctx context.Context, hb pkgmodel.Heartbeat) error
	ProcessPingResult(ctx context.Context, res pkgmodel.ResponsePing) error
	SyncServerLifecycle(ctx context.Context, event pkgmodel.ServerEvent, action pkgmodel.ServerActionType) error
}
