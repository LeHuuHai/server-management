package service

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/LeHuuHai/server-management/internal/domain/repo"
	apperr "github.com/LeHuuHai/server-management/internal/error"
	"github.com/LeHuuHai/server-management/internal/file/export"
	"github.com/LeHuuHai/server-management/internal/model"
)

type ServerService struct {
	repo     repo.ServerRepositoryInterface
	exporter export.Exporter
}

func (s *ServerService) CreateServer(ctx context.Context, server *model.Server) error {
	ip := net.ParseIP(server.IPv4)
	if ip == nil || ip.To4() == nil {
		return apperr.ErrInvalidIP
	}
	return s.repo.Create(ctx, server)
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
	ip := net.ParseIP(server.IPv4)
	if ip == nil || ip.To4() == nil {
		return nil, apperr.ErrInvalidIP
	}
	fields := map[string]any{}
	if server.ServerName != "" {
		fields["server_name"] = server.ServerName
	}
	if server.IPv4 != "" {
		fields["ipv4"] = server.IPv4
	}
	fields["metadata_updated_at"] = time.Now()
	return s.repo.Update(ctx, server.ServerID, fields)
}

func (s *ServerService) DeleteServer(ctx context.Context, serverID string) error {
	return s.repo.Delete(ctx, serverID)
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
	res.Failed = append(res.Failed, invalid...)
	res.FailedCnt += len(invalid)
	return res, nil
}

func (s *ServerService) ExportServer(ctx context.Context, filter model.ListServerFilter, exporter export.Exporter, writer io.Writer) error {
	servers, err := s.ListServer(ctx, filter)
	if err != nil {
		return err
	}
	return exporter.Export(ctx, writer, servers.Servers)
}

func NewServerService(r repo.ServerRepositoryInterface, e export.Exporter) *ServerService {
	return &ServerService{
		repo:     r,
		exporter: e,
	}
}
