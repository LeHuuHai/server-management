package handler

import (
	"context"
	"log/slog"
	"time"

	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/api"
	"github.com/LeHuuHai/server-management/microservices/heartbeat-gateway/internal/domain/service"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type HeartbeatHandler struct {
	svc service.GwServiceInterface
}

func NewHeartbeatHandler(svc service.GwServiceInterface) api.StrictServerInterface {
	return &HeartbeatHandler{
		svc: svc,
	}
}

// SendHeartbeat implements api.StrictServerInterface.
func (h *HeartbeatHandler) SendHeartbeat(ctx context.Context, request api.SendHeartbeatRequestObject) (api.SendHeartbeatResponseObject, error) {
	if request.Body == nil || request.Body.ServerId == "" {
		msg := "server_id is required"
		code := "400"
		return api.SendHeartbeat400JSONResponse{
			BadRequestJSONResponse: api.BadRequestJSONResponse{
				Message: &msg,
				Code:    &code,
			},
		}, nil
	}

	heartbeat := pkgmodel.Heartbeat{
		ServerID:  request.Body.ServerId,
		Timestamp: time.Now().UTC(),
	}

	slog.Info("Publishing heartbeat from strict handler", slog.String("server_id", heartbeat.ServerID))

	err := h.svc.PublishHeartbeat(ctx, heartbeat)
	if err != nil {
		slog.Error("Failed to publish heartbeat", slog.String("server_id", heartbeat.ServerID), slog.Any("error", err))
		msg := err.Error()
		code := "500"
		return api.SendHeartbeat500JSONResponse{
			InternalErrorJSONResponse: api.InternalErrorJSONResponse{
				Message: &msg,
				Code:    &code,
			},
		}, nil
	}

	slog.Info("Heartbeat accepted", slog.String("server_id", heartbeat.ServerID))
	return api.SendHeartbeat202Response{}, nil
}
