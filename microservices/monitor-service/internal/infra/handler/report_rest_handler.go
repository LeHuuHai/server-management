package handler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/api"
	domainservice "github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/service"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/model"
	apperr "github.com/LeHuuHai/server-management/microservices/pkg/apperr"
)

type ReportRestHandler struct {
	reportSvc domainservice.ReportServiceInterface
}

func NewReportRestHandler(reportSvc domainservice.ReportServiceInterface) *ReportRestHandler {
	return &ReportRestHandler{
		reportSvc: reportSvc,
	}
}

// Generate server report
// (POST /servers/report)
func (handler *ReportRestHandler) GenerateServerReport(ctx context.Context, request api.GenerateServerReportRequestObject) (api.GenerateServerReportResponseObject, error) {
	receivers := make([]string, len(request.Body.Receivers))
	for i, r := range request.Body.Receivers {
		receivers[i] = string(r)
	}

	req := model.GenServerReportRequest{
		From:      request.Body.From,
		To:        request.Body.To,
		Receivers: receivers,
	}

	slog.Info("handler: generate server report",
		slog.Any("from", req.From),
		slog.Any("to", req.To),
		slog.Any("receivers", req.Receivers),
	)

	err := handler.reportSvc.ReportServer(ctx, req)
	if err != nil {
		if errors.Is(err, apperr.ErrInvalidTimeRange) || errors.Is(err, apperr.ErrInvalidEmail) {
			slog.Warn("invalid request", slog.Any("err", err))
			return api.GenerateServerReport400JSONResponse{
				BadRequestJSONResponse: BadRequest(err),
			}, nil
		}
		slog.Error("failed to generate server report", slog.Any("err", err))
		return api.GenerateServerReport500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}

	slog.Info("handler: generate server report accepted",
		slog.Any("receivers", req.Receivers),
	)

	return api.GenerateServerReport202Response{}, nil
}
