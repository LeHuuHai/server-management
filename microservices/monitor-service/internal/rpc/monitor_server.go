package rpc

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	domainservice "github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/service"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/model"
	pb "github.com/LeHuuHai/server-management/microservices/pkg/pb/monitor"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MonitorHandler struct {
	pb.UnimplementedMonitorServiceServer
	reportSvc domainservice.ReportServiceInterface
}

func NewMonitorHandler(reportSvc domainservice.ReportServiceInterface) *MonitorHandler {
	return &MonitorHandler{
		reportSvc: reportSvc,
	}
}

func (h *MonitorHandler) DownloadReport(req *pb.DownloadReportRequest, stream pb.MonitorService_DownloadReportServer) error {
	if req.Filename == "" {
		return status.Error(codes.InvalidArgument, "filename is required")
	}

	filePath := fmt.Sprintf("./tmp/%s", req.Filename)
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return status.Errorf(codes.NotFound, "file %s not found", req.Filename)
		}
		return status.Errorf(codes.Internal, "failed to open file: %v", err)
	}
	defer file.Close()

	// 64 KB buffer chunk
	buf := make([]byte, 64*1024)

	for {
		n, err := file.Read(buf)
		if n > 0 {
			sendErr := stream.Send(&pb.DownloadReportResponse{
				ChunkData: buf[:n],
			})
			if sendErr != nil {
				return status.Errorf(codes.Unavailable, "failed to send chunk: %v", sendErr)
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return status.Errorf(codes.Internal, "failed to read file chunk: %v", err)
		}
	}

	return nil
}

func (h *MonitorHandler) GenerateReport(ctx context.Context, req *pb.GenerateReportRequest) (*pb.GenerateReportResponse, error) {
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

