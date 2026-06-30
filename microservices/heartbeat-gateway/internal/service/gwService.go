package service

import (
	"context"

	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/internal/domain/mq"
	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/internal/domain/service"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type GwService struct {
	publisher mq.PublisherInterface
}

func NewGwService(publisher mq.PublisherInterface) service.GwServiceInterface {
	return &GwService{
		publisher: publisher,
	}
}

func (s *GwService) PublishHeartbeat(ctx context.Context, heartbeat pkgmodel.Heartbeat) error {
	return s.publisher.Publish(ctx, heartbeat)
}
