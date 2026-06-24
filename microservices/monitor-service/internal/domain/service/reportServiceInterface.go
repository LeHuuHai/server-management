package service

import (
	"context"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/model"
)

type ReportServiceInterface interface {
	ReportServer(ctx context.Context, request model.GenServerReportRequest) error
}
