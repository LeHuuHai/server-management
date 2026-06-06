package handler

import (
	"context"
	"log/slog"
	"time"

	gwapi "github.com/LeHuuHai/server-management/api/gw"
	serviceinterface "github.com/LeHuuHai/server-management/internal/domain/service"
	"github.com/LeHuuHai/server-management/internal/middleware"
	"github.com/LeHuuHai/server-management/internal/model"
)

// impl StrictServerInterface
type GwHandler struct {
	GwService serviceinterface.GWServiceInterface
}

func NewGwHandler(gwService serviceinterface.GWServiceInterface) *GwHandler {
	return &GwHandler{
		GwService: gwService,
	}
}

// Send server heartbeat
// (POST /heartbeat)
func (h *GwHandler) SendHeartbeat(ctx context.Context, request gwapi.SendHeartbeatRequestObject) (gwapi.SendHeartbeatResponseObject, error) {
	logger := middleware.LoggerFromContext(ctx)

	heartbeat := model.Heartbeat{
		ServerID:  request.Body.ServerId,
		Timestamp: time.Now().UTC(),
	}

	logger.Info("handler: send heartbeat", slog.String("server_id", heartbeat.ServerID))

	err := h.GwService.PublishHeartbeat(ctx, heartbeat)
	if err != nil {
		logger.Error("failed to publish heartbeat", slog.String("server_id", heartbeat.ServerID), slog.Any("err", err))
		msg := err.Error()
		code := "500"
		return gwapi.SendHeartbeat500JSONResponse{
			InternalErrorJSONResponse: gwapi.InternalErrorJSONResponse{Message: &msg, Code: &code},
		}, nil
	}

	logger.Info("handler: send heartbeat accepted", slog.String("server_id", heartbeat.ServerID))

	return gwapi.SendHeartbeat202Response{}, nil
}
