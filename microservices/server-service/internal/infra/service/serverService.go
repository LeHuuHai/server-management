package service

import (
	"context"
	"net"
	"time"

	apperr "github.com/LeHuuHai/server-management/microservices/pkg/apperr"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/domain/publisher"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/domain/repo"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/model"
)

type ServerService struct {
	repo           repo.ServerRepositoryInterface
	eventPublisher publisher.EventPublisherInterface
}

func (s *ServerService) CreateServer(ctx context.Context, server *model.Server) (*model.Server, error) {
	ip := net.ParseIP(server.IPv4)
	if ip == nil || ip.To4() == nil {
		return nil, apperr.ErrInvalidIP
	}
	err := s.repo.Create(ctx, server)
	if err != nil {
		return nil, err
	}

	// publish event
	_ = s.eventPublisher.PublishServerCreated(ctx, server)

	return server, nil
}

func (s *ServerService) ListServer(ctx context.Context, filter model.ListServerFilter) (*model.ListServerResult, error) {
	// sorting
	switch filter.SortField {
	case model.SortByName,
		model.SortByCreatedAt:
	default:
		return nil, apperr.ErrInvalidSort
	}
	// pagination
	if filter.To-filter.From <= 0 || filter.From < 0 || filter.To <= 0 {
		return nil, apperr.ErrInvalidPagination
	}
	if filter.To-filter.From > 100 {
		filter.To = filter.From + 100
	}

	res, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *ServerService) UpdateServer(ctx context.Context, server *model.Server) (*model.Server, error) {
	fields := map[string]any{}
	if server.ServerName != "" {
		fields["server_name"] = server.ServerName
	}
	if server.IPv4 != "" {
		ip := net.ParseIP(server.IPv4)
		if ip == nil || ip.To4() == nil {
			return nil, apperr.ErrInvalidIP
		}
		fields["ipv4"] = server.IPv4
	}
	fields["metadata_updated_at"] = time.Now()
	newServer, err := s.repo.Update(ctx, server.ServerID, fields)
	if err != nil {
		return nil, err
	}

	// publish event
	_ = s.eventPublisher.PublishServerUpdated(ctx, newServer)

	return newServer, nil
}

func (s *ServerService) DeleteServer(ctx context.Context, serverID string) error {
	err := s.repo.Delete(ctx, serverID)
	if err != nil {
		return err
	}

	// publish event
	_ = s.eventPublisher.PublishServerDeleted(ctx, serverID)

	return nil
}

func (s *ServerService) ImportServer(ctx context.Context, serversData []model.ServerImport) (*model.CreateBatchServerResult, error) {
	invalid := make([]string, 0)
	valid := make([]model.Server, 0)
	for _, item := range serversData {
		ip := net.ParseIP(item.IPv4)
		if ip == nil || ip.To4() == nil {
			invalid = append(invalid, item.ServerID)
			continue
		}

		valid = append(valid, model.Server{
			ServerID:   item.ServerID,
			ServerName: item.ServerName,
			IPv4:       item.IPv4,
		})
	}

	res, err := s.repo.CreateBatch(ctx, valid)
	if err != nil {
		return nil, err
	}

	// publish event
	isSuccess := make(map[string]bool)
	for _, serverID := range res.Success {
		isSuccess[serverID] = true
	}
	for _, data := range serversData {
		success, ok := isSuccess[data.ServerID]
		if success && ok {
			_ = s.eventPublisher.PublishServerCreated(ctx, &model.Server{
				ServerID:   data.ServerID,
				ServerName: data.ServerName,
				IPv4:       data.IPv4,
			})
		}
	}

	res.Failed = append(res.Failed, invalid...)
	res.FailedCnt += len(invalid)

	return res, nil
}

func NewServerService(
	r repo.ServerRepositoryInterface,
	p publisher.EventPublisherInterface,
) *ServerService {
	return &ServerService{
		repo:           r,
		eventPublisher: p,
	}
}
