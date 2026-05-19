package domain

import (
	"context"

	"github.com/LeHuuHai/server-management/internal/model"
)

type BatchResult struct {
	Success    []string
	Failed     []string
	SuccessCnt int
	FailedCnt  int
}

type ServerRepository interface {
	Create(ctx context.Context, s *model.Server) error

	Update(ctx context.Context, id string, fields map[string]any) error

	Delete(ctx context.Context, id string) error

	List(ctx context.Context, filter model.ListServerFilter) ([]model.Server, int, error)

	CreateBatch(ctx context.Context, servers []model.Server) (*BatchResult, error)
}
