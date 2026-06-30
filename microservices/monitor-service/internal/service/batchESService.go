package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/repo"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/model"
)

type BatchESService struct {
	input    chan model.StatusLog
	maxSize  int
	timeout  time.Duration
	logsRepo repo.LogsRepositoryInterface[model.StatusLog]
}

func NewBatchESService(input chan model.StatusLog, size int, timeout time.Duration, logsRepo repo.LogsRepositoryInterface[model.StatusLog]) *BatchESService {
	return &BatchESService{
		input:    input,
		maxSize:  size,
		timeout:  timeout,
		logsRepo: logsRepo,
	}
}

func (s *BatchESService) Run(ctx context.Context) {
	timer := time.NewTicker(s.timeout)
	defer timer.Stop()
	buffer := make([]model.StatusLog, 0, s.maxSize)

	flush := func() {
		if len(buffer) == 0 {
			return
		}
		tmp := make([]model.StatusLog, len(buffer))
		copy(tmp, buffer)
		buffer = buffer[:0]

		go func(data []model.StatusLog) {
			err := s.logsRepo.WriteBatch(context.Background(), data)
			if err != nil {
				slog.Error("Failed to bulk write logs to Elasticsearch", "err", err)
			}
		}(tmp)
		slog.Info("Flushed logs batch to Elasticsearch", "batch_size", len(tmp))
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
			buffer = append(buffer, item)
			if len(buffer) >= s.maxSize {
				flush()
			}
		}
	}
}
