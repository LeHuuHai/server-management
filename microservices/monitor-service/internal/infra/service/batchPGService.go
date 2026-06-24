package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/repo"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/model"
)

type BatchPGService struct {
	input   chan model.LiveStatus
	maxSize int
	timeout time.Duration
	repo    repo.LiveStatusRepositoryInterface
}

func NewBatchPGService(input chan model.LiveStatus, size int, timeout time.Duration, repo repo.LiveStatusRepositoryInterface) *BatchPGService {
	return &BatchPGService{
		input:   input,
		maxSize: size,
		timeout: timeout,
		repo:    repo,
	}
}

func (s *BatchPGService) Run(ctx context.Context) {
	timer := time.NewTicker(s.timeout)
	defer timer.Stop()
	buffer := make(map[string]model.LiveStatus)

	flush := func() {
		if len(buffer) == 0 {
			return
		}

		// Convert map to slice for bulk update
		list := make([]model.LiveStatus, 0, len(buffer))
		for _, v := range buffer {
			list = append(list, v)
		}

		buffer = make(map[string]model.LiveStatus, s.maxSize)

		go func(data []model.LiveStatus) {
			err := s.repo.BulkUpdateLiveStatus(context.Background(), data)
			if err != nil {
				slog.Error("Failed to bulk update LiveStatus to Postgres", "err", err)
			}
		}(list)
		slog.Info("Flushed status batch to Postgres", "batch_size", len(list))
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case <-timer.C:
			flush()
		case item, ok := <-s.input:
			if !ok {
				flush()
				return
			}
			buffer[item.ServerID] = item
			if len(buffer) >= s.maxSize {
				flush()
			}
		}
	}
}
