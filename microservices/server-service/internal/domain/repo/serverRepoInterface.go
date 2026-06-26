package repo

import (
	"context"

	"github.com/LeHuuHai/server-management/microservices/server-service/internal/model"
)

type ServerRepositoryInterface interface {
	Create(ctx context.Context, s *model.ServerProfile) error

	Update(ctx context.Context, id string, fields map[string]any) (*model.ServerProfile, error)

	Delete(ctx context.Context, id string) error

	List(ctx context.Context, filter model.ListServerFilter) (*model.ListServerResult, error)

	CreateBatch(ctx context.Context, servers []model.ServerProfile) (*model.CreateBatchServerResult, error)
}
