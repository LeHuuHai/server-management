package worker

import (
	"context"
	"log/slog"
	"sync"

	"github.com/LeHuuHai/server-management/microservices/ping-worker/internal/domain/mq"
	"github.com/LeHuuHai/server-management/microservices/ping-worker/internal/domain/service"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type PingWorkerPool struct {
	numThreads int
	consumer   mq.PingRequestConsumerInterface
	publisher  mq.PingResponsePublisherInterface
	service    service.PingServiceInterface
}

func NewPingWorkerPool(numThreads int, consumer mq.PingRequestConsumerInterface, publisher mq.PingResponsePublisherInterface, svc service.PingServiceInterface) *PingWorkerPool {
	return &PingWorkerPool{
		numThreads: numThreads,
		consumer:   consumer,
		publisher:  publisher,
		service:    svc,
	}
}

func (p *PingWorkerPool) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	jobs := make(chan struct {
		Req        pkgmodel.RequestPing
		CommitFunc func(context.Context) error
	}, p.numThreads*2)

	var poolWg sync.WaitGroup
	poolWg.Add(p.numThreads + 1)

	// Reader Goroutine
	go func() {
		defer poolWg.Done()
		defer close(jobs)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				req, commitFunc, err := p.consumer.Read(ctx)
				if err != nil {
					slog.Warn("Failed to read ping request", slog.Any("error", err))
					continue
				}

				select {
				case jobs <- struct {
					Req        pkgmodel.RequestPing
					CommitFunc func(context.Context) error
				}{req, commitFunc}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// Worker Goroutines
	for i := 0; i < p.numThreads; i++ {
		go func() {
			defer poolWg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-jobs:
					if !ok {
						return
					}

					res, err := p.service.Ping(ctx, job.Req)
					if err != nil {
						slog.Warn("Ping service failed", slog.String("server_id", job.Req.ServerID), slog.Any("error", err))
						continue
					}

					err = p.publisher.Publish(ctx, res)
					if err != nil {
						slog.Warn("Failed to publish ping response", slog.String("server_id", job.Req.ServerID), slog.Any("error", err))
						continue
					}

					// Commit message after successful processing
					if job.CommitFunc != nil {
						err = job.CommitFunc(ctx)
						if err != nil {
							slog.Warn("Failed to commit ping request", slog.String("server_id", job.Req.ServerID), slog.Any("error", err))
						}
					}
				}
			}
		}()
	}

	poolWg.Wait()
}
