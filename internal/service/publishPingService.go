package service

import (
	"context"

	"github.com/LeHuuHai/server-management/internal/domain/mq"
)

type PublishPingService struct {
	publisher mq.Publisher
}

func (p *PublishPingService) PublishRequestPing(ctx context.Context, req []byte) error {
	err := p.publisher.Publish(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func NewPublishPingService(p mq.Publisher) *PublishPingService {
	return &PublishPingService{
		publisher: p,
	}
}
