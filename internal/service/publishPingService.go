package service

import (
	"context"

	"github.com/LeHuuHai/server-management/internal/domain/kafka"
)

type PublishPingService struct {
	publisher *kafka.Publisher
}

func (p *PublishPingService) PublishRequestPing(ctx context.Context, req []byte) error {
	err := p.publisher.Publish(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func NewPublishPingService(p *kafka.Publisher) *PublishPingService {
	return &PublishPingService{
		publisher: p,
	}
}
