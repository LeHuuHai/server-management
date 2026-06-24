package serviceinterface

import (
	"context"

	"github.com/LeHuuHai/server-management/microservices/server-service/internal/model"
)

type ServerServiceInterface interface {
	CreateServer(ctx context.Context, server *model.ServerProfile) (*model.ServerProfile, error)

	ListServer(ctx context.Context, filter model.ListServerFilter) (*model.ListServerResult, error)

	UpdateServer(ctx context.Context, server *model.ServerProfile) (*model.ServerProfile, error)

	DeleteServer(ctx context.Context, serverID string) error

	ImportServer(ctx context.Context, serversData []model.ServerImport) (*model.CreateBatchServerResult, error)
}
