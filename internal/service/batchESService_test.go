package service_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/LeHuuHai/server-management/internal/service"
)

func captureESFlush() (func([]model.ServerEvent) error, func() [][]model.ServerEvent) {
	var mu sync.Mutex
	var batches [][]model.ServerEvent

	flush := func(data []model.ServerEvent) error {
		cp := make([]model.ServerEvent, len(data))
		copy(cp, data)
		mu.Lock()
		batches = append(batches, cp)
		mu.Unlock()
		return nil
	}
	get := func() [][]model.ServerEvent {
		mu.Lock()
		defer mu.Unlock()
		return batches
	}
	return flush, get
}

func totalESItems(batches [][]model.ServerEvent) int {
	n := 0
	for _, b := range batches {
		n += len(b)
	}
	return n
}

func TestBatchES_FlushOnMaxSize(t *testing.T) {
	flush, getBatches := captureESFlush()
	input := make(chan model.ServerEvent, 10)

	svc := service.NewBatchESService(input, 3, 10*time.Second, flush)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Run(ctx)

	input <- model.ServerEvent{ServerID: "a"}
	input <- model.ServerEvent{ServerID: "b"}
	input <- model.ServerEvent{ServerID: "c"}

	time.Sleep(50 * time.Millisecond)

	batches := getBatches()
	if len(batches) == 0 {
		t.Fatal("expected at least one flush, got none")
	}
	if totalESItems(batches) != 3 {
		t.Errorf("expected 3 items flushed, got %d", totalESItems(batches))
	}
}

func TestBatchES_FlushOnTimeout(t *testing.T) {
	flush, getBatches := captureESFlush()
	input := make(chan model.ServerEvent, 10)

	svc := service.NewBatchESService(input, 100, 80*time.Millisecond, flush)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Run(ctx)

	input <- model.ServerEvent{ServerID: "x"}
	input <- model.ServerEvent{ServerID: "y"}

	time.Sleep(200 * time.Millisecond)

	batches := getBatches()
	if len(batches) == 0 {
		t.Fatal("expected flush on timeout, got none")
	}
	if totalESItems(batches) != 2 {
		t.Errorf("expected 2 items flushed, got %d", totalESItems(batches))
	}
}

func TestBatchES_FlushOnContextCancel(t *testing.T) {
	flush, getBatches := captureESFlush()
	input := make(chan model.ServerEvent, 10)

	svc := service.NewBatchESService(input, 100, 10*time.Second, flush)
	ctx, cancel := context.WithCancel(context.Background())

	go svc.Run(ctx)

	input <- model.ServerEvent{ServerID: "a"}
	input <- model.ServerEvent{ServerID: "b"}

	time.Sleep(20 * time.Millisecond)
	cancel()
	time.Sleep(50 * time.Millisecond)

	batches := getBatches()
	if len(batches) == 0 {
		t.Fatal("expected flush on context cancel, got none")
	}
	if totalESItems(batches) != 2 {
		t.Errorf("expected 2 items, got %d", totalESItems(batches))
	}
}

func TestBatchES_FlushOnChannelClose(t *testing.T) {
	flush, getBatches := captureESFlush()
	input := make(chan model.ServerEvent, 10)

	svc := service.NewBatchESService(input, 100, 10*time.Second, flush)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Run(ctx)

	input <- model.ServerEvent{ServerID: "s1"}
	input <- model.ServerEvent{ServerID: "s2"}

	time.Sleep(20 * time.Millisecond)
	close(input)
	time.Sleep(50 * time.Millisecond)

	batches := getBatches()
	if len(batches) == 0 {
		t.Fatal("expected flush on channel close, got none")
	}
	if totalESItems(batches) != 2 {
		t.Errorf("expected 2 items, got %d", totalESItems(batches))
	}
}

func TestBatchES_NoFlushWhenBufferEmpty(t *testing.T) {
	flush, getBatches := captureESFlush()
	input := make(chan model.ServerEvent, 10)

	svc := service.NewBatchESService(input, 10, 50*time.Millisecond, flush)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Run(ctx)

	time.Sleep(150 * time.Millisecond)

	if len(getBatches()) != 0 {
		t.Error("expected no flush for empty buffer, but flush was called")
	}
}

// TestBatchES_DuplicateServerID kiểm tra ESService KHÔNG dedup
// (khác PGService dùng map — ES giữ tất cả event kể cả trùng ServerID)
func TestBatchES_DuplicateServerID(t *testing.T) {
	flush, getBatches := captureESFlush()
	input := make(chan model.ServerEvent, 10)

	svc := service.NewBatchESService(input, 100, 80*time.Millisecond, flush)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Run(ctx)

	input <- model.ServerEvent{ServerID: "dup"}
	input <- model.ServerEvent{ServerID: "dup"}
	input <- model.ServerEvent{ServerID: "dup"}

	time.Sleep(200 * time.Millisecond)

	batches := getBatches()
	if len(batches) == 0 {
		t.Fatal("no batch received")
	}
	if totalESItems(batches) != 3 {
		t.Errorf("expected 3 items (no dedup), got %d", totalESItems(batches))
	}
}

func TestBatchES_MultipleFlushCycles(t *testing.T) {
	flush, getBatches := captureESFlush()
	input := make(chan model.ServerEvent, 20)

	svc := service.NewBatchESService(input, 3, 10*time.Second, flush)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Run(ctx)

	for i := 0; i < 6; i++ {
		input <- model.ServerEvent{ServerID: string(rune('a' + i))}
	}

	time.Sleep(100 * time.Millisecond)

	batches := getBatches()
	if len(batches) < 2 {
		t.Errorf("expected at least 2 batches, got %d", len(batches))
	}
	if totalESItems(batches) != 6 {
		t.Errorf("expected 6 total items, got %d", totalESItems(batches))
	}
}
