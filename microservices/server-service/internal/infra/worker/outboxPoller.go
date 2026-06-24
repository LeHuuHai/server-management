package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/LeHuuHai/server-management/microservices/server-service/internal/domain/publisher"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/domain/repo"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/model"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type OutboxPoller struct {
	outboxRepo repo.OutboxRepositoryInterface
	publisher  publisher.EventPublisherInterface
}

func NewOutboxPoller(r repo.OutboxRepositoryInterface, p publisher.EventPublisherInterface) *OutboxPoller {
	return &OutboxPoller{
		outboxRepo: r,
		publisher:  p,
	}
}

func (p *OutboxPoller) Start(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Outbox poller stopped")
			return
		case <-ticker.C:
			p.processBatch(ctx)
		}
	}
}

func (p *OutboxPoller) processBatch(ctx context.Context) {
	events, err := p.outboxRepo.FetchPendingEvents(ctx, 100)
	if err != nil {
		slog.Error("Failed to fetch outbox events", "error", err)
		return
	}

	if len(events) == 0 {
		return
	}

	var successfulIDs []string

	for _, ev := range events {
		var serverEvent pkgmodel.ServerEvent
		if err := json.Unmarshal(ev.Payload, &serverEvent); err != nil {
			slog.Error("Failed to unmarshal outbox payload", "error", err, "id", ev.ID)
			continue
		}

		var pubErr error
		switch ev.Topic {
		case "server_created":
			pubErr = p.publisher.PublishServerCreated(ctx, &model.ServerProfile{
				ServerID:   serverEvent.ServerID,
				ServerName: serverEvent.ServerName,
				IPv4:       serverEvent.IPv4,
			})
		case "server_updated":
			pubErr = p.publisher.PublishServerUpdated(ctx, &model.ServerProfile{
				ServerID:   serverEvent.ServerID,
				ServerName: serverEvent.ServerName,
				IPv4:       serverEvent.IPv4,
			})
		case "server_deleted":
			pubErr = p.publisher.PublishServerDeleted(ctx, serverEvent.ServerID)
		default:
			slog.Warn("Unknown topic in outbox event", "topic", ev.Topic)
			continue
		}

		if pubErr != nil {
			slog.Error("Failed to publish outbox event", "error", pubErr, "id", ev.ID)
			// Stop processing this batch to preserve ordering (optional) or continue
			continue
		}

		successfulIDs = append(successfulIDs, ev.ID)
	}

	if len(successfulIDs) > 0 {
		if err := p.outboxRepo.DeleteEvents(ctx, successfulIDs); err != nil {
			slog.Error("Failed to delete published outbox events", "error", err)
		}
	}
}
