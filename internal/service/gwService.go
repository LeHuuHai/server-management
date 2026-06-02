package service

import (
	"context"
	"encoding/json"

	"github.com/LeHuuHai/server-management/internal/domain/mq"
	"github.com/LeHuuHai/server-management/internal/model"
)

type GwService struct {
	publisher      mq.Publisher
	heartbeatTopic string
}

func NewGwService(publisher mq.Publisher, heartbeatTopic string) *GwService {
	return &GwService{
		publisher:      publisher,
		heartbeatTopic: heartbeatTopic,
	}
}

func (s *GwService) PublishHeartbeat(ctx context.Context, heartbeat model.Heartbeat) error {
	value, err := json.Marshal(heartbeat)
	if err != nil {
		return err
	}
	return s.publisher.Publish(ctx, mq.Message{
		Topic: s.heartbeatTopic,
		Value: value,
	})
}
