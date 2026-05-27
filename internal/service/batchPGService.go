package service

import (
	"context"
	"time"

	"github.com/LeHuuHai/server-management/internal/model"
)

type BatchPGService struct {
	Input     chan model.Server
	MaxSize   int
	Timeout   time.Duration
	FlushFunc func(map[string]model.Server) error
}

func NewBatchPGService(input chan model.Server, size int, timeout time.Duration, flushFunc func(map[string]model.Server) error) *BatchPGService {
	return &BatchPGService{
		Input:     input,
		MaxSize:   size,
		Timeout:   timeout,
		FlushFunc: flushFunc,
	}
}

func (s *BatchPGService) Run(ctx context.Context) {
	timer := time.NewTicker(s.Timeout)
	defer timer.Stop()
	buffer := make(map[string]model.Server)

	f := func() {
		if len(buffer) == 0 {
			return
		}
		tmp := buffer
		buffer = make(map[string]model.Server, s.MaxSize)
		go func(data map[string]model.Server) {
			_ = s.FlushFunc(data)
		}(tmp)
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
			buffer[item.ServerID] = item
			if len(buffer) >= s.MaxSize {
				f()
			}
		}
	}
}
