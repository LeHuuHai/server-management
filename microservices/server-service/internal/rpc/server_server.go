package rpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	// "google.golang.org/grpc/metadata"

	serverpb "github.com/LeHuuHai/server-management/microservices/pkg/pb/server"
	service "github.com/LeHuuHai/server-management/microservices/server-service/internal/domain/service"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/model"
)

type ServerHandler struct {
	serverpb.UnimplementedServerServiceServer
	svc service.ServerServiceInterface
}

func NewServerHandler(svc service.ServerServiceInterface) *ServerHandler {
	return &ServerHandler{
		svc: svc,
	}
}

func mapModelToServerPb(m *model.ServerProfile) *serverpb.ServerProfile {
	return &serverpb.ServerProfile{
		ServerId:   m.ServerID,
		ServerName: m.ServerName,
		Ipv4:       m.IPv4,
		CreatedAt:  m.CreatedAt.Unix(),
		UpdatedAt:  m.UpdatedAt.Unix(),
	}
}

func (h *ServerHandler) CreateServer(ctx context.Context, req *serverpb.CreateServerRequest) (*serverpb.ServerProfile, error) {
	// md, ok := metadata.FromIncomingContext(ctx)
	// if ok { ... } // Extract User info if needed

	m := &model.ServerProfile{
		ServerID:   req.ServerId,
		ServerName: req.ServerName,
		IPv4:       req.Ipv4,
	}
	res, err := h.svc.CreateServer(ctx, m)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create server: %v", err)
	}

	return mapModelToServerPb(res), nil
}

func (h *ServerHandler) UpdateServer(ctx context.Context, req *serverpb.UpdateServerRequest) (*serverpb.ServerProfile, error) {
	m := &model.ServerProfile{
		ServerID:   req.ServerId,
		ServerName: req.ServerName,
		IPv4:       req.Ipv4,
	}
	res, err := h.svc.UpdateServer(ctx, m)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update server: %v", err)
	}

	return mapModelToServerPb(res), nil
}

func (h *ServerHandler) DeleteServer(ctx context.Context, req *serverpb.DeleteServerRequest) (*serverpb.DeleteServerResponse, error) {
	err := h.svc.DeleteServer(ctx, req.ServerId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete server: %v", err)
	}
	return &serverpb.DeleteServerResponse{}, nil
}

func (h *ServerHandler) ListServers(ctx context.Context, req *serverpb.ListServersRequest) (*serverpb.ListServersResponse, error) {
	filter := model.ListServerFilter{
		From:      int(req.From),
		To:        int(req.To),
		SortField: model.ServerSortField(req.SortField),
		Desc:      req.Desc,
	}

	res, err := h.svc.ListServer(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list servers: %v", err)
	}

	var pbServers []*serverpb.ServerProfile
	for _, s := range res.Servers {
		sCopy := s // prevent pointer loop issue
		pbServers = append(pbServers, mapModelToServerPb(&sCopy))
	}

	return &serverpb.ListServersResponse{
		Servers: pbServers,
		Total:   int32(res.Total),
	}, nil
}

func (h *ServerHandler) ImportServers(ctx context.Context, req *serverpb.ImportServersRequest) (*serverpb.ImportServersResponse, error) {
	// 1. Dùng req.FileContent (bytes array của file xlsx)
	// 2. Chuyển cho FileService deserialize -> []model.ServerImport
	// (Note: Cần thêm method Deserialize cho FileService hoặc thực hiện ở đây,
	// nhưng do cấu trúc chưa lộ rõ nên tạm mock logic gọi service)

	// mock calling ImportServer
	// imports := deserialize(req.FileContent)
	// res, err := h.svc.ImportServer(ctx, imports)

	return nil, status.Errorf(codes.Unimplemented, "ImportServers not implemented fully yet")
}

func (h *ServerHandler) ExportServers(ctx context.Context, req *serverpb.ExportServersRequest) (*serverpb.ExportServersResponse, error) {
	// filter := model.ListServerFilter{
	// 	From:      int(req.From),
	// 	To:        int(req.To),
	// 	SortField: model.ServerSortField(req.SortField),
	// 	Desc:      req.Desc,
	// }
	// servers, _ := h.svc.ListServer(ctx, filter)
	// bytes := serializeToXlsx(servers)
	return nil, status.Errorf(codes.Unimplemented, "ExportServers not implemented fully yet")
}
