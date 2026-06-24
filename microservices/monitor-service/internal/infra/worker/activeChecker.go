package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/mq"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/repo"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/model"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type ActiveChecker struct {
	monitoredServerRepo repo.MonitoredServerRepositoryInterface
	liveStatusRepo      repo.LiveStatusRepositoryInterface
	publisher           mq.PingRequestPublisherInterface
	cyclePing           time.Duration
	heartbeatTimeout    time.Duration
}

func NewActiveChecker(
	monitoredServerRepo repo.MonitoredServerRepositoryInterface,
	liveStatusRepo repo.LiveStatusRepositoryInterface,
	publisher mq.PingRequestPublisherInterface,
	cyclePingMs int,
	heartbeatTimeoutMs int,
) *ActiveChecker {
	return &ActiveChecker{
		monitoredServerRepo: monitoredServerRepo,
		liveStatusRepo:      liveStatusRepo,
		publisher:           publisher,
		cyclePing:           time.Duration(cyclePingMs) * time.Millisecond,
		heartbeatTimeout:    time.Duration(heartbeatTimeoutMs) * time.Millisecond,
	}
}

func (c *ActiveChecker) Start(ctx context.Context) {
	ticker := time.NewTicker(c.cyclePing)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			start := time.Now()
			servers, err := c.monitoredServerRepo.List(ctx)
			if err != nil {
				slog.Error("ActiveChecker failed to list monitored servers from Postgres", "err", err)
				continue
			}

			statuses, err := c.liveStatusRepo.List(ctx)
			if err != nil {
				slog.Error("ActiveChecker failed to list live statuses from Postgres", "err", err)
				continue
			}

			statusMap := make(map[string]model.LiveStatus)
			for _, s := range statuses {
				statusMap[s.ServerID] = s
			}

			cnt := 0
			for _, item := range servers {
				statusInfo, exists := statusMap[item.ServerID]

				// If server has sent heartbeat recently (within heartbeatTimeout), skip active ping check
				if exists && statusInfo.LastHeartbeatAt != nil && time.Since(*statusInfo.LastHeartbeatAt) < c.heartbeatTimeout {
					continue
				}

				cnt++
				req := pkgmodel.RequestPing{
					ServerID:   item.ServerID,
					ServerName: item.ServerName,
					IP:         item.IPv4,
				}

				err = c.publisher.Publish(ctx, req)
				if err != nil {
					slog.Warn("Publish RequestPing failed", slog.Any("req", req), slog.Any("err", err))
					continue
				}
			}

			elapse := time.Since(start)
			slog.Info("ActiveChecker check cycle completed", "published_pings", cnt, "elapse", elapse)
		}
	}
}
