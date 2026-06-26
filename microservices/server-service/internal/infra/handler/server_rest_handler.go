package handler

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/LeHuuHai/server-management/microservices/pkg/apperr"
	"github.com/LeHuuHai/server-management/microservices/server-service/api"
	serviceinterface "github.com/LeHuuHai/server-management/microservices/server-service/internal/domain/service"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/model"
)

type ServerRestHandler struct {
	service serviceinterface.ServerServiceInterface
}

func NewServerRestHandler(s serviceinterface.ServerServiceInterface) *ServerRestHandler {
	return &ServerRestHandler{
		service: s,
	}
}

func strPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// Get list servers
// (GET /servers)
func (handler *ServerRestHandler) GetListServers(ctx context.Context, request api.GetListServersRequestObject) (api.GetListServersResponseObject, error) {
	params := request.Params
	filter := model.ListServerFilter{
		From:      int(params.From),
		To:        int(params.To),
		SortField: model.ServerSortField(params.SortField),
		Desc:      params.Desc,
	}

	res, err := handler.service.ListServer(ctx, filter)
	if err != nil {
		if errors.Is(err, apperr.ErrInvalidSort) || errors.Is(err, apperr.ErrInvalidPagination) {
			return api.GetListServers400JSONResponse{
				BadRequestJSONResponse: BadRequest(err),
			}, nil
		}
		return api.GetListServers500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}

	items := make([]api.Server, len(res.Servers))
	for idx, s := range res.Servers {
		items[idx] = api.Server{
			ServerId:   s.ServerID,
			ServerName: s.ServerName,
			Ipv4:       s.IPv4,
			CreatedAt:  timePtr(s.CreatedAt),
			UpdatedAt:  timePtr(s.UpdatedAt),
		}
	}
	total := int(res.Total)
	return api.GetListServers200JSONResponse{
		Items: &items,
		Total: &total,
	}, nil
}

// Create server
// (POST /servers)
func (handler *ServerRestHandler) CreateServer(ctx context.Context, request api.CreateServerRequestObject) (api.CreateServerResponseObject, error) {
	server := model.ServerAddress{
		ServerID:   request.Body.ServerId,
		ServerName: request.Body.ServerName,
		IPv4:       request.Body.Ipv4,
	}

	newServer, err := handler.service.CreateServer(ctx, &server)
	if err != nil {
		if errors.Is(err, apperr.ErrDuplicateServer) {
			return api.CreateServer409JSONResponse{
				ConflictJSONResponse: Conflict(err),
			}, nil
		}
		if errors.Is(err, apperr.ErrInvalidIP) {
			return api.CreateServer400JSONResponse{
				BadRequestJSONResponse: BadRequest(err),
			}, nil
		}
		return api.CreateServer500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}

	return api.CreateServer201JSONResponse{
		ServerId:   newServer.ServerID,
		ServerName: newServer.ServerName,
		Ipv4:       newServer.IPv4,
		CreatedAt:  timePtr(newServer.CreatedAt),
		UpdatedAt:  timePtr(newServer.UpdatedAt),
	}, nil
}

// Update server
// (PATCH /servers/{server_id})
func (handler *ServerRestHandler) UpdateServer(ctx context.Context, request api.UpdateServerRequestObject) (api.UpdateServerResponseObject, error) {
	server := model.ServerAddress{
		ServerID:   request.ServerId,
		ServerName: *request.Body.ServerName,
		IPv4:       *request.Body.Ipv4,
	}

	s, err := handler.service.UpdateServer(ctx, &server)
	if err != nil {
		if errors.Is(err, apperr.ErrRecordNotFound) {
			return api.UpdateServer404JSONResponse{
				NotFoundJSONResponse: NotFound(err),
			}, nil
		}
		return api.UpdateServer500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}

	return api.UpdateServer200JSONResponse{
		ServerId:   s.ServerID,
		ServerName: s.ServerName,
		Ipv4:       s.IPv4,
		CreatedAt:  timePtr(s.CreatedAt),
		UpdatedAt:  timePtr(s.UpdatedAt),
	}, nil
}

// Delete server
// (DELETE /servers/{server_id})
func (handler *ServerRestHandler) DeleteServer(ctx context.Context, request api.DeleteServerRequestObject) (api.DeleteServerResponseObject, error) {
	if err := handler.service.DeleteServer(ctx, request.ServerId); err != nil {
		if errors.Is(err, apperr.ErrRecordNotFound) {
			return api.DeleteServer404JSONResponse{
				NotFoundJSONResponse: NotFound(err),
			}, nil
		}
		return api.DeleteServer500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}

	return api.DeleteServer204Response{}, nil
}

// Export servers
// (GET /servers/export)
func (handler *ServerRestHandler) ExportServers(ctx context.Context, request api.ExportServersRequestObject) (api.ExportServersResponseObject, error) {
	slog.Warn("handler: export not implemented yet")
	return api.ExportServers500JSONResponse{
		InternalErrorJSONResponse: InternalError(errors.New("export not fully implemented yet")),
	}, nil
}

// Import server
// (POST /servers/import)
func (handler *ServerRestHandler) ImportServer(ctx context.Context, request api.ImportServerRequestObject) (api.ImportServerResponseObject, error) {
	slog.Warn("handler: import not implemented yet")
	return api.ImportServer500JSONResponse{
		InternalErrorJSONResponse: InternalError(errors.New("import not fully implemented yet")),
	}, nil
}
