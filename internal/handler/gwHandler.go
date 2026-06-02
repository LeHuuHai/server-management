package handler

import (
	"context"
	"time"

	gwapi "github.com/LeHuuHai/server-management/api/gw"
	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/LeHuuHai/server-management/internal/service"
)

// impl StrictServerInterface
type GwHandler struct {
	GwService *service.GwService
}

func NewGwHandler(gwService *service.GwService) *GwHandler {
	return &GwHandler{
		GwService: gwService,
	}
}

// Send server heartbeat
// (POST /heartbeat)
func (h *GwHandler) SendHeartbeat(ctx context.Context, request gwapi.SendHeartbeatRequestObject) (gwapi.SendHeartbeatResponseObject, error) {
	heartbeat := model.Heartbeat{
		ServerID:  request.Body.ServerId,
		Timestamp: time.Now().UTC(),
	}
	err := h.GwService.PublishHeartbeat(ctx, heartbeat)
	if err != nil {
		msg := err.Error()
		code := "500"
		return gwapi.SendHeartbeat500JSONResponse{
			InternalErrorJSONResponse: gwapi.InternalErrorJSONResponse{Message: &msg, Code: &code},
		}, nil
	}
	return gwapi.SendHeartbeat202Response{}, nil
}
