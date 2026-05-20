package service

import (
	"context"
	"encoding/json"
	"net"

	"github.com/LeHuuHai/server-management/internal/domain/kafka"
	apperr "github.com/LeHuuHai/server-management/internal/error"
	"github.com/LeHuuHai/server-management/internal/model"
)

type PublishService struct {
	publisher *kafka.Publisher
}

func (p *PublishService) PublishRequestPing(ctx context.Context, req model.RequestPing) error {
	ip := net.ParseIP(req.IP)
	if ip == nil || ip.To4() == nil {
		return apperr.ErrInvalidIP
	}
	reqByte, err := json.Marshal(req)
	if err != nil {
		return err
	}
	err = p.publisher.Publish(ctx, reqByte)
	if err != nil {
		return err
	}
	return nil
}

func NewPublishService(p *kafka.Publisher) *PublishService {
	return &PublishService{
		publisher: p,
	}
}
