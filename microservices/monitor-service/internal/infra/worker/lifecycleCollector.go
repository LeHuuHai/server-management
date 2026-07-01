package worker

import (
	"context"
	"log/slog"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/mq"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/service"
)

type LifecycleConsumer struct {
	consumer mq.LifecycleConsumerInterface
	svc      service.MonitorServiceInterface
}

func NewLifecycleConsumer(consumer mq.LifecycleConsumerInterface, svc service.MonitorServiceInterface) *LifecycleConsumer {
	return &LifecycleConsumer{
		consumer: consumer,
		svc:      svc,
	}
}

func (c *LifecycleConsumer) Start(ctx context.Context) {
	for {
		event, action, commitFunc, err := c.consumer.Read(ctx)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				slog.Error("LifecycleConsumer failed to read message", "err", err)
				continue
			}
		}
		slog.Info("Received server lifecycle event", "action", action, "server_id", event.ServerID, "server_name", event.ServerName, "ipv4", event.IPv4, "version", event.Version)
		err = c.svc.SyncServerLifecycle(ctx, event, action)
		if err != nil {
			slog.Error("Failed to sync server lifecycle event", "action", action, "server_id", event.ServerID, "err", err)
			continue
		}

		if err := commitFunc(ctx); err != nil {
			slog.Error("Failed to commit lifecycle message", "server_id", event.ServerID, "err", err)
		} else {
			slog.Info("Successfully synced server lifecycle event", "action", action, "server_id", event.ServerID)
		}
	}
}
