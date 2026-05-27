package service

import (
	"context"
	"time"

	"github.com/LeHuuHai/server-management/internal/model"
)

type BatchESService struct {
	Input     chan model.ResponsePing
	MaxSize   int
	Timeout   time.Duration
	FlushFunc func([]model.ResponsePing) error
}

func NewBatchESService(input chan model.ResponsePing, size int, timeout time.Duration, flushFunc func([]model.ResponsePing) error) *BatchESService {
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
	buffer := make([]model.ResponsePing, 0, s.MaxSize)

	f := func() {
		if len(buffer) == 0 {
			return
		}
		tmp := make([]model.ResponsePing, len(buffer))
		copy(tmp, buffer)
		buffer = buffer[:0]
		go func(data []model.ResponsePing) {
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
			buffer = append(buffer, item)
			if len(buffer) >= s.MaxSize {
				f()
			}
		}
	}
}
