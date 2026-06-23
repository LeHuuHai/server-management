package publisher

import (
	"context"

	"github.com/LeHuuHai/server-management/microservices/server-service/internal/model"
)

type EventPublisherInterface interface {
	PublishServerCreated(ctx context.Context, server *model.Server) error
	PublishServerUpdated(ctx context.Context, server *model.Server) error
	PublishServerDeleted(ctx context.Context, serverID string) error
}
