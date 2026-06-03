package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/LeHuuHai/server-management/internal/model"
)

type BatchESService struct {
	Input     chan model.ServerEvent
	MaxSize   int
	Timeout   time.Duration
	FlushFunc func([]model.ServerEvent) error
}

func NewBatchESService(input chan model.ServerEvent, size int, timeout time.Duration, flushFunc func([]model.ServerEvent) error) *BatchESService {
	return &BatchESService{
		Input:     input,
		MaxSize:   size,
		Timeout:   timeout,
		FlushFunc: flushFunc,
	}
}

func (s *BatchESService) Run(ctx context.Context) {
	timer := time.NewTicker(s.Timeout)
	defer timer.Stop()
	buffer := make([]model.ServerEvent, 0, s.MaxSize)

	f := func() {
		if len(buffer) == 0 {
			return
		}
		tmp := make([]model.ServerEvent, len(buffer))
		copy(tmp, buffer)
		buffer = buffer[:0]
		go func(data []model.ServerEvent) {
			_ = s.FlushFunc(data)
		}(tmp)
		slog.Info("Flushing batch to ES", "batch_size", len(tmp))
	}

	for {
		select {
		case <-ctx.Done():
			f()
			return
		case <-timer.C:
			f()
		case item, ok := <-s.Input:
			if !ok {
				f()
				return
			}
			buffer = append(buffer, item)
			if len(buffer) >= s.MaxSize {
				f()
			}
		}
	}
}
