package repo

import (
	"context"

	"github.com/LeHuuHai/server-management/microservices/server-service/internal/model"
)

type ServerRepositoryInterface interface {
	Create(ctx context.Context, s *model.Server) error

	Update(ctx context.Context, id string, fields map[string]any) (*model.Server, error)

	Delete(ctx context.Context, id string) error

	List(ctx context.Context, filter model.ListServerFilter) (*model.ListServerResult, error)

	CreateBatch(ctx context.Context, servers []model.Server) (*model.CreateBatchServerResult, error)

	AllMetadata(ctx context.Context) ([]model.ServerMetadata, error)

	BulkUpdateServers(ctx context.Context, items []model.Server) error
}
