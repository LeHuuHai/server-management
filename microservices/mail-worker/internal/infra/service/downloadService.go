package service

import (
	"bytes"
	"context"
	"io"

	"github.com/LeHuuHai/server-management/microservices/mail-worker/internal/domain/service"
	monitorpb "github.com/LeHuuHai/server-management/microservices/pkg/pb/monitor"
	"google.golang.org/grpc"
)

type DownloadService struct {
	client monitorpb.InternalFileTransferServiceClient
}

func NewDownloadService(conn grpc.ClientConnInterface) service.DownloadServiceInterface {
	return &DownloadService{
		client: monitorpb.NewInternalFileTransferServiceClient(conn),
	}
}

func (s *DownloadService) Download(ctx context.Context, filename string) ([]byte, error) {
	req := &monitorpb.DownloadReportRequest{
		Filename: filename,
	}

	stream, err := s.client.DownloadReport(ctx, req)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		buffer.Write(res.ChunkData)
	}

	return buffer.Bytes(), nil
}
