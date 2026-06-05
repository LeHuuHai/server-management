package service_test

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/LeHuuHai/server-management/internal/service"
)

// helper: tạo flushFunc ghi lại các batch đã flush
func captureFlush() (func([]model.ServerMetadata) error, func() [][]model.ServerMetadata) {
	var mu sync.Mutex
	var batches [][]model.ServerMetadata

	flush := func(data []model.ServerMetadata) error {
		mu.Lock()
		defer mu.Unlock()
		cp := make([]model.ServerMetadata, len(data))
		copy(cp, data)
		batches = append(batches, cp)
		return nil
	}
	get := func() [][]model.ServerMetadata {
		mu.Lock()
		defer mu.Unlock()
		return batches
	}
	return flush, get
}

// helper: đếm tổng số item trong tất cả batches
func totalItems(batches [][]model.ServerMetadata) int {
	n := 0
	for _, b := range batches {
		n += len(b)
	}
	return n
}

// TestBatch_FlushOnMaxSize kiểm tra flush khi buffer đạt MaxSize
func TestBatch_FlushOnMaxSize(t *testing.T) {
	flush, getBatches := captureFlush()
	input := make(chan model.ServerMetadata, 10)

	svc := service.NewBatchServerMetadataService(input, 3, 10*time.Second, flush)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Run(ctx)

	// Gửi đúng MaxSize items
	for i := 0; i < 3; i++ {
		input <- model.ServerMetadata{ServerID: strconv.Itoa(i)}
	}

	// Đợi goroutine flush xong
	time.Sleep(50 * time.Millisecond)

	batches := getBatches()
	if len(batches) == 0 {
		t.Fatal("expected at least one flush, got none")
	}
	if totalItems(batches) != 3 {
		t.Errorf("expected 3 items flushed, got %d", totalItems(batches))
	}
}

// TestBatch_FlushOnTimeout kiểm tra flush khi hết timeout dù buffer chưa đầy
func TestBatch_FlushOnTimeout(t *testing.T) {
	flush, getBatches := captureFlush()
	input := make(chan model.ServerMetadata, 10)

	svc := service.NewBatchServerMetadataService(input, 100, 80*time.Millisecond, flush)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Run(ctx)

	input <- model.ServerMetadata{ServerID: "1"}
	input <- model.ServerMetadata{ServerID: "2"}

	// Đợi timeout trigger flush
	time.Sleep(200 * time.Millisecond)

	batches := getBatches()
	if len(batches) == 0 {
		t.Fatal("expected flush on timeout, got none")
	}
	if totalItems(batches) != 2 {
		t.Errorf("expected 2 items flushed, got %d", totalItems(batches))
	}
}

// TestBatch_FlushOnContextCancel kiểm tra flush khi context bị cancel
func TestBatch_FlushOnContextCancel(t *testing.T) {
	flush, getBatches := captureFlush()
	input := make(chan model.ServerMetadata, 10)

	svc := service.NewBatchServerMetadataService(input, 100, 10*time.Second, flush)
	ctx, cancel := context.WithCancel(context.Background())

	go svc.Run(ctx)

	input <- model.ServerMetadata{ServerID: "1"}
	input <- model.ServerMetadata{ServerID: "2"}
	input <- model.ServerMetadata{ServerID: "3"}

	time.Sleep(20 * time.Millisecond)
	cancel() // trigger ctx.Done()
	time.Sleep(50 * time.Millisecond)

	batches := getBatches()
	if len(batches) == 0 {
		t.Fatal("expected flush on context cancel, got none")
	}
	if totalItems(batches) != 3 {
		t.Errorf("expected 3 items, got %d", totalItems(batches))
	}
}

// TestBatch_FlushOnChannelClose kiểm tra flush khi channel bị đóng
func TestBatch_FlushOnChannelClose(t *testing.T) {
	flush, getBatches := captureFlush()
	input := make(chan model.ServerMetadata, 10)

	svc := service.NewBatchServerMetadataService(input, 100, 10*time.Second, flush)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Run(ctx)

	input <- model.ServerMetadata{ServerID: "1"}
	input <- model.ServerMetadata{ServerID: "2"}

	time.Sleep(20 * time.Millisecond)
	close(input)
	time.Sleep(50 * time.Millisecond)

	batches := getBatches()
	if len(batches) == 0 {
		t.Fatal("expected flush on channel close, got none")
	}
	if totalItems(batches) != 2 {
		t.Errorf("expected 2 items, got %d", totalItems(batches))
	}
}

// TestBatch_NoFlushWhenBufferEmpty đảm bảo không flush khi buffer rỗng
func TestBatch_NoFlushWhenBufferEmpty(t *testing.T) {
	flush, getBatches := captureFlush()
	input := make(chan model.ServerMetadata, 10)

	svc := service.NewBatchServerMetadataService(input, 10, 50*time.Millisecond, flush)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Run(ctx)

	// Không gửi item nào, chờ timeout trigger
	time.Sleep(150 * time.Millisecond)

	if len(getBatches()) != 0 {
		t.Error("expected no flush for empty buffer, but flush was called")
	}
}

// TestBatch_MultipleFlushCycles kiểm tra nhiều chu kỳ flush liên tiếp
func TestBatch_MultipleFlushCycles(t *testing.T) {
	flush, getBatches := captureFlush()
	input := make(chan model.ServerMetadata, 20)

	svc := service.NewBatchServerMetadataService(input, 3, 10*time.Second, flush)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Run(ctx)

	// Gửi 6 items → nên tạo ra 2 batch (mỗi batch 3 items)
	for i := 0; i < 6; i++ {
		input <- model.ServerMetadata{ServerID: strconv.Itoa(i)}
	}

	time.Sleep(100 * time.Millisecond)

	batches := getBatches()
	if len(batches) < 2 {
		t.Errorf("expected at least 2 batches, got %d", len(batches))
	}
	if totalItems(batches) != 6 {
		t.Errorf("expected 6 total items, got %d", totalItems(batches))
	}
}

// TestBatch_DataIntegrity kiểm tra dữ liệu trong batch không bị mutate sau flush
func TestBatch_DataIntegrity(t *testing.T) {
	flush, getBatches := captureFlush()
	input := make(chan model.ServerMetadata, 10)

	svc := service.NewBatchServerMetadataService(input, 2, 10*time.Second, flush)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Run(ctx)

	input <- model.ServerMetadata{ServerID: "1"}
	input <- model.ServerMetadata{ServerID: "2"}

	time.Sleep(80 * time.Millisecond)

	batches := getBatches()
	if len(batches) == 0 {
		t.Fatal("no batch received")
	}

	batch := batches[0]
	ids := map[string]bool{}
	for _, m := range batch {
		ids[m.ServerID] = true
	}
	if !ids["1"] || !ids["2"] {
		t.Errorf("expected ServerIDs q and w, got %+v", batch)
	}
}
