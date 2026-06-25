package rpc

import (
	"context"
	"time"

	domainservice "github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/service"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/model"
	pb "github.com/LeHuuHai/server-management/microservices/pkg/pb/monitor"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ReportHandler struct {
	pb.UnimplementedReportManagementServiceServer
	reportSvc domainservice.ReportServiceInterface
}

func NewReportHandler(reportSvc domainservice.ReportServiceInterface) *ReportHandler {
	return &ReportHandler{
		reportSvc: reportSvc,
	}
}

func (h *ReportHandler) GenerateReport(ctx context.Context, req *pb.GenerateReportRequest) (*pb.GenerateReportResponse, error) {
	if req.FromTimestamp > req.ToTimestamp {
		return nil, status.Error(codes.InvalidArgument, "invalid time range")
	}
	if len(req.Receivers) == 0 {
		return nil, status.Error(codes.InvalidArgument, "receivers list is required")
	}

	reportReq := model.GenServerReportRequest{
		From:      time.Unix(req.FromTimestamp, 0),
		To:        time.Unix(req.ToTimestamp, 0),
		Receivers: req.Receivers,
	}

	err := h.reportSvc.ReportServer(ctx, reportReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate report: %v", err)
	}

	return &pb.GenerateReportResponse{
		Success: true,
	}, nil
}
