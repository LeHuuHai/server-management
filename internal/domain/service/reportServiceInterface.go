package serviceinterface

import (
	"context"

	"github.com/LeHuuHai/server-management/internal/model"
)

type ReportServiceInterface interface {
	ReportServer(ctx context.Context, request model.GenServerReportRequest) error
}
