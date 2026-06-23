package rpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	// "google.golang.org/grpc/metadata"

	"github.com/LeHuuHai/server-management/microservices/pkg/pb/server"
	service "github.com/LeHuuHai/server-management/microservices/server-service/internal/domain/service"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/model"
)

type ServerHandler struct {
	server.UnimplementedServerServiceServer
	svc service.ServerServiceInterface
}

func NewServerHandler(svc service.ServerServiceInterface) *ServerHandler {
	return &ServerHandler{
		svc: svc,
	}
}

func mapModelToServerPb(m *model.Server) *server.Server {
	return &server.Server{
		ServerId:          m.ServerID,
		ServerName:        m.ServerName,
		Ipv4:              m.IPv4,
		Status:            string(m.Status),
		CreatedAt:         m.CreatedAt.Unix(),
		MetadataUpdatedAt: m.MetadataUpdatedAt.Unix(),
		LastPingAt:        m.LastPingAt.Unix(),
	}
}

func (h *ServerHandler) CreateServer(ctx context.Context, req *server.CreateServerRequest) (*server.Server, error) {
	// md, ok := metadata.FromIncomingContext(ctx)
	// if ok { ... } // Extract User info if needed

	m := &model.Server{
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

func (h *ServerHandler) UpdateServer(ctx context.Context, req *server.UpdateServerRequest) (*server.Server, error) {
	m := &model.Server{
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

func (h *ServerHandler) DeleteServer(ctx context.Context, req *server.DeleteServerRequest) (*server.DeleteServerResponse, error) {
	err := h.svc.DeleteServer(ctx, req.ServerId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete server: %v", err)
	}
	return &server.DeleteServerResponse{}, nil
}

func (h *ServerHandler) ListServers(ctx context.Context, req *server.ListServersRequest) (*server.ListServersResponse, error) {
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

	var pbServers []*server.Server
	for _, s := range res.Servers {
		sCopy := s // prevent pointer loop issue
		pbServers = append(pbServers, mapModelToServerPb(&sCopy))
	}

	return &server.ListServersResponse{
		Servers: pbServers,
		Total:   int32(res.Total),
	}, nil
}

func (h *ServerHandler) ImportServers(ctx context.Context, req *server.ImportServersRequest) (*server.ImportServersResponse, error) {
	// 1. Dùng req.FileContent (bytes array của file xlsx)
	// 2. Chuyển cho FileService deserialize -> []model.ServerImport
	// (Note: Cần thêm method Deserialize cho FileService hoặc thực hiện ở đây,
	// nhưng do cấu trúc chưa lộ rõ nên tạm mock logic gọi service)

	// mock calling ImportServer
	// imports := deserialize(req.FileContent)
	// res, err := h.svc.ImportServer(ctx, imports)

	return nil, status.Errorf(codes.Unimplemented, "ImportServers not implemented fully yet")
}

func (h *ServerHandler) ExportServers(ctx context.Context, req *server.ExportServersRequest) (*server.ExportServersResponse, error) {
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
