package publisher

import (
	"context"

	"github.com/LeHuuHai/server-management/microservices/server-service/internal/model"
)

type EventPublisherInterface interface {
	PublishServerCreated(ctx context.Context, server *model.ServerProfile) error
	PublishServerUpdated(ctx context.Context, server *model.ServerProfile) error
	PublishServerDeleted(ctx context.Context, serverID string) error
}
