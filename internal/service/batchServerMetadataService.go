package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/LeHuuHai/server-management/internal/model"
)

type BatchServerMetadataService struct {
	Input     chan model.ServerMetadata
	MaxSize   int
	Timeout   time.Duration
	FlushFunc func([]model.ServerMetadata) error
}

func NewBatchServerMetadataService(input chan model.ServerMetadata, size int, timeout time.Duration, flushFunc func([]model.ServerMetadata) error) *BatchServerMetadataService {
	return &BatchServerMetadataService{
		Input:     input,
		MaxSize:   size,
		Timeout:   timeout,
		FlushFunc: flushFunc,
	}
}

func (s *BatchServerMetadataService) Run(ctx context.Context) {
	timer := time.NewTicker(s.Timeout)
	defer timer.Stop()
	buffer := make([]model.ServerMetadata, 0, s.MaxSize)

	f := func() {
		if len(buffer) == 0 {
			return
		}
		tmp := make([]model.ServerMetadata, len(buffer))
		copy(tmp, buffer)
		buffer = buffer[:0]
		go func(data []model.ServerMetadata) {
			_ = s.FlushFunc(data)
		}(tmp)
		slog.Info("Flushing batch to immem", "batch_size", len(tmp))
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
