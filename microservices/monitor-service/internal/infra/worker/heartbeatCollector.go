package worker

import (
	"context"
	"log/slog"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/mq"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/service"
)

type HeartbeatConsumer struct {
	consumer mq.HeartbeatConsumerInterface
	svc      service.MonitorServiceInterface
}

func NewHeartbeatConsumer(consumer mq.HeartbeatConsumerInterface, svc service.MonitorServiceInterface) *HeartbeatConsumer {
	return &HeartbeatConsumer{
		consumer: consumer,
		svc:      svc,
	}
}

func (c *HeartbeatConsumer) Start(ctx context.Context) {
	for {
		hb, commitFunc, err := c.consumer.Read(ctx)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				slog.Error("HeartbeatConsumer failed to read message", "err", err)
				continue
			}
		}

		err = c.svc.ProcessHeartbeat(ctx, hb)
		if err != nil {
			slog.Error("Failed to process heartbeat", "server_id", hb.ServerID, "err", err)
			continue
		}

		if err := commitFunc(ctx); err != nil {
			slog.Error("Failed to commit heartbeat message", "server_id", hb.ServerID, "err", err)
		}
	}
}
