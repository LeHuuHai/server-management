package serviceinterface

import (
	"context"

	"github.com/LeHuuHai/server-management/microservices/server-service/internal/model"
)

type ServerServiceInterface interface {
	CreateServer(ctx context.Context, server *model.Server) (*model.Server, error)

	ListServer(ctx context.Context, filter model.ListServerFilter) (*model.ListServerResult, error)

	UpdateServer(ctx context.Context, server *model.Server) (*model.Server, error)

	DeleteServer(ctx context.Context, serverID string) error

	ImportServer(ctx context.Context, serversData []model.ServerImport) (*model.CreateBatchServerResult, error)
}
