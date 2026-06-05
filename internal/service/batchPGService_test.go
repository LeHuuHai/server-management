package service_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/LeHuuHai/server-management/internal/service"
)

func capturePGFlush() (func(map[string]model.Server) error, func() []map[string]model.Server) {
	var mu sync.Mutex
	var batches []map[string]model.Server

	flush := func(data map[string]model.Server) error {
		cp := make(map[string]model.Server, len(data))
		for k, v := range data {
			cp[k] = v
		}
		mu.Lock()
		batches = append(batches, cp)
		mu.Unlock()
		return nil
	}
	get := func() []map[string]model.Server {
		mu.Lock()
		defer mu.Unlock()
		return batches
	}
	return flush, get
}

func totalPGItems(batches []map[string]model.Server) int {
	n := 0
	for _, b := range batches {
		n += len(b)
	}
	return n
}

func TestBatchPG_FlushOnMaxSize(t *testing.T) {
	flush, getBatches := capturePGFlush()
	input := make(chan model.Server, 10)

	svc := service.NewBatchPGService(input, 3, 10*time.Second, flush)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Run(ctx)

	input <- model.Server{ServerID: "a"}
	input <- model.Server{ServerID: "b"}
	input <- model.Server{ServerID: "c"}

	time.Sleep(50 * time.Millisecond)

	batches := getBatches()
	if len(batches) == 0 {
		t.Fatal("expected at least one flush, got none")
	}
	if totalPGItems(batches) != 3 {
		t.Errorf("expected 3 items flushed, got %d", totalPGItems(batches))
	}
}

func TestBatchPG_FlushOnTimeout(t *testing.T) {
	flush, getBatches := capturePGFlush()
	input := make(chan model.Server, 10)

	svc := service.NewBatchPGService(input, 100, 80*time.Millisecond, flush)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Run(ctx)

	input <- model.Server{ServerID: "x"}
	input <- model.Server{ServerID: "y"}

	time.Sleep(200 * time.Millisecond)

	batches := getBatches()
	if len(batches) == 0 {
		t.Fatal("expected flush on timeout, got none")
	}
	if totalPGItems(batches) != 2 {
		t.Errorf("expected 2 items flushed, got %d", totalPGItems(batches))
	}
}

func TestBatchPG_FlushOnContextCancel(t *testing.T) {
	flush, getBatches := capturePGFlush()
	input := make(chan model.Server, 10)

	svc := service.NewBatchPGService(input, 100, 10*time.Second, flush)
	ctx, cancel := context.WithCancel(context.Background())

	go svc.Run(ctx)

	input <- model.Server{ServerID: "a"}
	input <- model.Server{ServerID: "b"}

	time.Sleep(20 * time.Millisecond)
	cancel()
	time.Sleep(50 * time.Millisecond)

	batches := getBatches()
	if len(batches) == 0 {
		t.Fatal("expected flush on context cancel, got none")
	}
	if totalPGItems(batches) != 2 {
		t.Errorf("expected 2 items, got %d", totalPGItems(batches))
	}
}

func TestBatchPG_FlushOnChannelClose(t *testing.T) {
	flush, getBatches := capturePGFlush()
	input := make(chan model.Server, 10)

	svc := service.NewBatchPGService(input, 100, 10*time.Second, flush)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Run(ctx)

	input <- model.Server{ServerID: "s1"}
	input <- model.Server{ServerID: "s2"}

	time.Sleep(20 * time.Millisecond)
	close(input)
	time.Sleep(50 * time.Millisecond)

	batches := getBatches()
	if len(batches) == 0 {
		t.Fatal("expected flush on channel close, got none")
	}
	if totalPGItems(batches) != 2 {
		t.Errorf("expected 2 items, got %d", totalPGItems(batches))
	}
}

func TestBatchPG_NoFlushWhenBufferEmpty(t *testing.T) {
	flush, getBatches := capturePGFlush()
	input := make(chan model.Server, 10)

	svc := service.NewBatchPGService(input, 10, 50*time.Millisecond, flush)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Run(ctx)

	time.Sleep(150 * time.Millisecond)

	if len(getBatches()) != 0 {
		t.Error("expected no flush for empty buffer, but flush was called")
	}
}

// TestBatchPG_DedupByServerID kiểm tra đặc thù của PGService:
// map dedup theo ServerID, item sau ghi đè item trước
func TestBatchPG_DedupByServerID(t *testing.T) {
	flush, getBatches := capturePGFlush()
	input := make(chan model.Server, 10)

	svc := service.NewBatchPGService(input, 100, 80*time.Millisecond, flush)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Run(ctx)

	input <- model.Server{ServerID: "dup", ServerName: "first"}
	input <- model.Server{ServerID: "dup", ServerName: "second"}
	input <- model.Server{ServerID: "other"}

	time.Sleep(200 * time.Millisecond)

	batches := getBatches()
	if len(batches) == 0 {
		t.Fatal("no batch received")
	}

	batch := batches[0]
	if len(batch) != 2 {
		t.Errorf("expected 2 unique keys, got %d", len(batch))
	}
	if batch["dup"].ServerName != "second" {
		t.Errorf("expected last write wins, got name=%q", batch["dup"].ServerName)
	}
}

func TestBatchPG_MultipleFlushCycles(t *testing.T) {
	flush, getBatches := capturePGFlush()
	input := make(chan model.Server, 20)

	svc := service.NewBatchPGService(input, 3, 10*time.Second, flush)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Run(ctx)

	for i := 0; i < 6; i++ {
		input <- model.Server{ServerID: string(rune('a' + i))}
	}

	time.Sleep(100 * time.Millisecond)

	batches := getBatches()
	if len(batches) < 2 {
		t.Errorf("expected at least 2 batches, got %d", len(batches))
	}
	if totalPGItems(batches) != 6 {
		t.Errorf("expected 6 total items, got %d", totalPGItems(batches))
	}
}
