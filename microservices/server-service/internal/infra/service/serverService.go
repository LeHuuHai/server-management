package service

import (
	"context"
	"encoding/json"
	"net"
	"time"

	apperr "github.com/LeHuuHai/server-management/microservices/pkg/apperr"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/domain/repo"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/model"
	"github.com/google/uuid"
)

type ServerService struct {
	txManager  repo.TxManagerInterface
	repo       repo.ServerRepositoryInterface
	outboxRepo repo.OutboxRepositoryInterface
}

func (s *ServerService) CreateServer(ctx context.Context, server *model.Server) (*model.Server, error) {
	ip := net.ParseIP(server.IPv4)
	if ip == nil || ip.To4() == nil {
		return nil, apperr.ErrInvalidIP
	}

	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Create(txCtx, server); err != nil {
			return err
		}

		payload, _ := json.Marshal(model.ServerEvent{
			ServerID:   server.ServerID,
			ServerName: server.ServerName,
			IPv4:       server.IPv4,
			Timestamp:  time.Now(),
		})

		return s.outboxRepo.CreateEvent(txCtx, &model.OutboxEvent{
			ID:        uuid.New().String(),
			Topic:     "server_created",
			Payload:   payload,
			Status:    model.OutboxStatusPending,
			CreatedAt: time.Now(),
		})
	})

	if err != nil {
		return nil, err
	}

	return server, nil
}

func (s *ServerService) ListServer(ctx context.Context, filter model.ListServerFilter) (*model.ListServerResult, error) {
	switch filter.SortField {
	case model.SortByName, model.SortByCreatedAt:
	default:
		return nil, apperr.ErrInvalidSort
	}

	if filter.To-filter.From <= 0 || filter.From < 0 || filter.To <= 0 {
		return nil, apperr.ErrInvalidPagination
	}
	if filter.To-filter.From > 100 {
		filter.To = filter.From + 100
	}

	return s.repo.List(ctx, filter)
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

	var newServer *model.Server

	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		var err error
		newServer, err = s.repo.Update(txCtx, server.ServerID, fields)
		if err != nil {
			return err
		}

		payload, _ := json.Marshal(model.ServerEvent{
			ServerID:   newServer.ServerID,
			ServerName: newServer.ServerName,
			IPv4:       newServer.IPv4,
			Timestamp:  time.Now(),
		})

		return s.outboxRepo.CreateEvent(txCtx, &model.OutboxEvent{
			ID:        uuid.New().String(),
			Topic:     "server_updated",
			Payload:   payload,
			Status:    model.OutboxStatusPending,
			CreatedAt: time.Now(),
		})
	})

	if err != nil {
		return nil, err
	}

	return newServer, nil
}

func (s *ServerService) DeleteServer(ctx context.Context, serverID string) error {
	return s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Delete(txCtx, serverID); err != nil {
			return err
		}

		payload, _ := json.Marshal(model.ServerEvent{
			ServerID:  serverID,
			Timestamp: time.Now(),
		})

		return s.outboxRepo.CreateEvent(txCtx, &model.OutboxEvent{
			ID:        uuid.New().String(),
			Topic:     "server_deleted",
			Payload:   payload,
			Status:    model.OutboxStatusPending,
			CreatedAt: time.Now(),
		})
	})
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

	var res *model.CreateBatchServerResult

	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		var err error
		res, err = s.repo.CreateBatch(txCtx, valid)
		if err != nil {
			return err
		}

		isSuccess := make(map[string]bool)
		for _, serverID := range res.Success {
			isSuccess[serverID] = true
		}

		for _, data := range serversData {
			if isSuccess[data.ServerID] {
				payload, _ := json.Marshal(model.ServerEvent{
					ServerID:   data.ServerID,
					ServerName: data.ServerName,
					IPv4:       data.IPv4,
					Timestamp:  time.Now(),
				})

				err := s.outboxRepo.CreateEvent(txCtx, &model.OutboxEvent{
					ID:        uuid.New().String(),
					Topic:     "server_created",
					Payload:   payload,
					Status:    model.OutboxStatusPending,
					CreatedAt: time.Now(),
				})
				if err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	res.Failed = append(res.Failed, invalid...)
	res.FailedCnt += len(invalid)

	return res, nil
}

func NewServerService(
	tm repo.TxManagerInterface,
	r repo.ServerRepositoryInterface,
	outbox repo.OutboxRepositoryInterface,
) *ServerService {
	return &ServerService{
		txManager:  tm,
		repo:       r,
		outboxRepo: outbox,
	}
}
