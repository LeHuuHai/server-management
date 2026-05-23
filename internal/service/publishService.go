package service

import (
	"context"

	"github.com/LeHuuHai/server-management/internal/domain/mq"
)

type PublishService struct {
	publisher mq.Publisher
}

func (p *PublishService) Publish(ctx context.Context, topic string, value []byte) error {
	msg := mq.Message{
		Topic: topic,
		Value: value,
	}
	err := p.publisher.Publish(ctx, msg)
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
