package service

import (
	"context"

	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/internal/domain/mq"
	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/internal/domain/service"
	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/internal/model"
)

type GwService struct {
	publisher mq.PublisherInterface
}

func NewGwService(publisher mq.PublisherInterface) service.GwServiceInterface {
	return &GwService{
		publisher: publisher,
	}
}

func (s *GwService) PublishHeartbeat(ctx context.Context, heartbeat model.Heartbeat) error {
	return s.publisher.Publish(ctx, heartbeat)
}
