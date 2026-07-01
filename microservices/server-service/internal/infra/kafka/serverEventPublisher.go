package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/domain/publisher"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/model"
	"github.com/segmentio/kafka-go"
)

type serverEventPublisher struct {
	writer *kafka.Writer
}

func NewServerEventPublisher(w *kafka.Writer) publisher.EventPublisherInterface {
	return &serverEventPublisher{
		writer: w,
	}
}

func (p *serverEventPublisher) publish(ctx context.Context, eventType pkgmodel.ServerEventType, server *model.ServerProfile) error {
	event := pkgmodel.ServerEvent{
		ServerID:   server.ServerID,
		ServerName: server.ServerName,
		IPv4:       server.IPv4,
		Timestamp:  time.Now(),
		Version:    server.Version,
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	slog.Info("Publishing event to Kafka", "eventType", eventType, "serverID", server.ServerID, "serverName", server.ServerName, "IPv4", server.IPv4, "Version", server.Version)
	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(server.ServerID),
		Value: bytes,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(eventType)},
		},
	})
	if err != nil {
		slog.Error("Failed to publish event to Kafka", "error", err)
	}
	return err
}

func (p *serverEventPublisher) PublishServerCreated(ctx context.Context, server *model.ServerProfile) error {
	return p.publish(ctx, pkgmodel.ServerCreateEvent, server)
}

func (p *serverEventPublisher) PublishServerUpdated(ctx context.Context, server *model.ServerProfile) error {
	return p.publish(ctx, pkgmodel.ServerUpdateEvent, server)
}

func (p *serverEventPublisher) PublishServerDeleted(ctx context.Context, serverID string) error {
	// For deleted, we might not have the full Server model readily available.
	// We'll create a dummy server object with just the ID.
	dummyServer := &model.ServerProfile{
		ServerID: serverID,
	}
	return p.publish(ctx, pkgmodel.ServerDeleteEvent, dummyServer)
}
