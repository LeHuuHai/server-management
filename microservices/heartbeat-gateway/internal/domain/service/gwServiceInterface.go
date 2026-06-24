package service

import (
	"context"

	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/internal/model"
)

type GwServiceInterface interface {
	PublishHeartbeat(ctx context.Context, heartbeat model.Heartbeat) error
}
