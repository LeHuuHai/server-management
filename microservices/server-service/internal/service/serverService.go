package service

import (
	"context"
	"encoding/json"
	"net"
	"time"

	apperr "github.com/LeHuuHai/server-management/microservices/pkg/apperr"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/domain/repo"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/model"
	"github.com/google/uuid"
)

type ServerService struct {
	txManager  repo.TxManagerInterface
	repo       repo.ServerRepositoryInterface
	outboxRepo repo.OutboxRepositoryInterface
}

func (s *ServerService) CreateServer(ctx context.Context, server *model.ServerAddress) (*model.ServerProfile, error) {
	ip := net.ParseIP(server.IPv4)
	if ip == nil || ip.To4() == nil {
		return nil, apperr.ErrInvalidIP
	}

	serverProfile := &model.ServerProfile{
		ServerID:   server.ServerID,
		ServerName: server.ServerName,
		IPv4:       server.IPv4,
	}
	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Create(txCtx, serverProfile); err != nil {
			return err
		}

		payload, _ := json.Marshal(pkgmodel.ServerEvent{
			ServerID:   server.ServerID,
			ServerName: server.ServerName,
			IPv4:       server.IPv4,
			Timestamp:  time.Now(),
			Version:    1,
		})

		return s.outboxRepo.CreateEvent(txCtx, &model.OutboxEvent{
			ID:        uuid.New().String(),
			Topic:     pkgmodel.ServerCreateEvent,
			Payload:   payload,
			Status:    model.OutboxStatusPending,
			CreatedAt: time.Now(),
		})
	})

	if err != nil {
		return nil, err
	}

	return serverProfile, nil
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

func (s *ServerService) UpdateServer(ctx context.Context, server *model.ServerAddress) (*model.ServerProfile, error) {
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
	fields["updated_at"] = time.Now()

	var newServer *model.ServerProfile

	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		var err error
		newServer, err = s.repo.Update(txCtx, server.ServerID, fields)
		if err != nil {
			return err
		}

		payload, _ := json.Marshal(pkgmodel.ServerEvent{
			ServerID:   newServer.ServerID,
			ServerName: newServer.ServerName,
			IPv4:       newServer.IPv4,
			Timestamp:  time.Now(),
			Version:    newServer.Version,
		})

		return s.outboxRepo.CreateEvent(txCtx, &model.OutboxEvent{
			ID:        uuid.New().String(),
			Topic:     pkgmodel.ServerUpdateEvent,
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

		payload, _ := json.Marshal(pkgmodel.ServerEvent{
			ServerID:  serverID,
			Timestamp: time.Now(),
		})

		return s.outboxRepo.CreateEvent(txCtx, &model.OutboxEvent{
			ID:        uuid.New().String(),
			Topic:     pkgmodel.ServerDeleteEvent,
			Payload:   payload,
			Status:    model.OutboxStatusPending,
			CreatedAt: time.Now(),
		})
	})
}

func (s *ServerService) ImportServer(ctx context.Context, serversData []model.ServerAddress) (*model.CreateBatchServerResult, error) {
	invalid := make([]string, 0)
	valid := make([]model.ServerProfile, 0)

	for _, item := range serversData {
		ip := net.ParseIP(item.IPv4)
		if ip == nil || ip.To4() == nil {
			invalid = append(invalid, item.ServerID)
			continue
		}
		valid = append(valid, model.ServerProfile{
			ServerID:   item.ServerID,
			ServerName: item.ServerName,
			IPv4:       item.IPv4,
		})
	}

	var res *model.CreateBatchServerResult = &model.CreateBatchServerResult{
		Success:    make([]string, 0),
		Failed:     make([]string, 0),
		SuccessCnt: 0,
		FailedCnt:  0,
	}

	for _, saddress := range serversData {
		_, err := s.CreateServer(ctx, &saddress)
		if err != nil {
			res.Failed = append(res.Failed, saddress.ServerID)
			res.FailedCnt++
		} else {
			res.Success = append(res.Success, saddress.ServerID)
			res.SuccessCnt++
		}
	}
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
