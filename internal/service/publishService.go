package service

import (
	"context"

	"github.com/LeHuuHai/server-management/internal/domain/mq"
)

type PublishService struct {
	publisher mq.Publisher
}

func (p *PublishService) Publish(ctx context.Context, topic string, msg []byte) error {
	err := p.publisher.Publish(ctx, topic, msg)
	if err != nil {
		return err
	}
	return nil
}

func NewPublishService(p mq.Publisher) *PublishService {
	return &PublishService{
		publisher: p,
	}
}
