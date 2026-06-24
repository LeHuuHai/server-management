package rpc

import (
	"fmt"
	"io"
	"os"

	pb "github.com/LeHuuHai/server-management/microservices/pkg/pb/monitor"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MonitorHandler struct {
	pb.UnimplementedMonitorServiceServer
}

func NewMonitorHandler() *MonitorHandler {
	return &MonitorHandler{}
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
