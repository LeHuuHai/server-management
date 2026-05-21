package cache

import (
	"context"

	"github.com/LeHuuHai/server-management/internal/model"
)

type ServerMetadataCacheInterface interface {
	Create(ctx context.Context, s model.ServerMetadata)

	Update(ctx context.Context, s model.ServerMetadata)

	Delete(ctx context.Context, s string)

	BatchCreate(ctx context.Context, s []model.ServerMetadata)

	List(ctx context.Context) []model.ServerMetadata
}
