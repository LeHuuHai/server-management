package rpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/internal/domain/service"
	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/internal/model"
	pb "github.com/LeHuuHai/server-management/microservices/pkg/pb/heartbeat"
)

type HeartbeatHandler struct {
	pb.UnimplementedHeartbeatServiceServer
	svc service.GwServiceInterface
}

func NewHeartbeatHandler(svc service.GwServiceInterface) *HeartbeatHandler {
	return &HeartbeatHandler{
		svc: svc,
	}
}

func (h *HeartbeatHandler) SendHeartbeat(ctx context.Context, req *pb.SendHeartbeatRequest) (*pb.SendHeartbeatResponse, error) {
	if req.ServerId == "" {
		return nil, status.Error(codes.InvalidArgument, "server_id is required")
	}

	timestamp := time.Unix(req.Timestamp, 0)
	if req.Timestamp == 0 {
		timestamp = time.Now()
	}

	hb := model.Heartbeat{
		ServerID:  req.ServerId,
		Timestamp: timestamp,
	}

	err := h.svc.PublishHeartbeat(ctx, hb)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish heartbeat: %v", err)
	}

	return &pb.SendHeartbeatResponse{
		Success: true,
		Message: "Heartbeat accepted",
	}, nil
}
