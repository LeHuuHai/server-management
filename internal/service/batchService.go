package service

import (
	"context"
	"time"
)

type BatchService[T any] struct {
	Input     chan T
	MaxSize   int
	Timeout   time.Duration
	FlushFunc func([]T) error
}

func NewBatchService[T any](input chan T, size int, timeout time.Duration, flushFunc func([]T) error) *BatchService[T] {
	return &BatchService[T]{
		Input:     input,
		MaxSize:   size,
		Timeout:   timeout,
		FlushFunc: flushFunc,
	}
}

func (s *BatchService[T]) Run(ctx context.Context) {
	timer := time.NewTicker(s.Timeout)
	defer timer.Stop()
	buffer := make([]T, 0, s.MaxSize)

	f := func() {
		if len(buffer) == 0 {
			return
		}
		tmp := make([]T, len(buffer))
		copy(tmp, buffer)
		buffer = buffer[:0]
		if err := s.FlushFunc(tmp); err != nil {
			// log error
		}
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
