package worker

import (
	"context"
	"log/slog"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/mq"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/service"
)

type PingResultConsumer struct {
	consumer mq.PingResponseConsumerInterface
	svc      service.MonitorServiceInterface
}

func NewPingResultConsumer(consumer mq.PingResponseConsumerInterface, svc service.MonitorServiceInterface) *PingResultConsumer {
	return &PingResultConsumer{
		consumer: consumer,
		svc:      svc,
	}
}

func (c *PingResultConsumer) Start(ctx context.Context) {
	for {
		res, commitFunc, err := c.consumer.Read(ctx)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				slog.Error("PingResultConsumer failed to read message", "err", err)
				continue
			}
		}

		err = c.svc.ProcessPingResult(ctx, res)
		if err != nil {
			slog.Error("Failed to process ping result", "server_id", res.ServerID, "err", err)
			continue
		}

		if err := commitFunc(ctx); err != nil {
			slog.Error("Failed to commit ping result message", "server_id", res.ServerID, "err", err)
		}
	}
}
