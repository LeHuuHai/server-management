package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/LeHuuHai/server-management/microservices/pkg/apperr"
	"github.com/LeHuuHai/server-management/microservices/server-service/api"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/domain/file/deserialize"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/domain/file/export"
	serviceinterface "github.com/LeHuuHai/server-management/microservices/server-service/internal/domain/service"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/model"
)

type ServerRestHandler struct {
	service  serviceinterface.ServerServiceInterface
	exporter export.ServerExporter
	importer deserialize.ServerDeserializer
}

func NewServerRestHandler(s serviceinterface.ServerServiceInterface, e export.ServerExporter, i deserialize.ServerDeserializer) *ServerRestHandler {
	return &ServerRestHandler{
		service:  s,
		importer: i,
		exporter: e,
	}
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
	params := request.Params
	filter := model.ListServerFilter{
		From:      params.From,
		To:        params.To,
		SortField: model.ServerSortField(params.SortField),
		Desc:      params.Desc,
	}

	slog.Info("handler: export servers", slog.Any("filter", filter))

	res, err := handler.service.ListServer(ctx, filter)
	if err != nil {
		if errors.Is(err, apperr.ErrInvalidSort) || errors.Is(err, apperr.ErrInvalidPagination) {
			slog.Warn("invalid request", slog.Any("err", err))
			return api.ExportServers400JSONResponse{
				BadRequestJSONResponse: BadRequest(err),
			}, nil
		}

		slog.Error("failed to list servers", slog.Any("err", err))

		return api.ExportServers500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}

	slog.Info("handler: exporting", slog.Int("total", res.Total), slog.String("file_type", handler.exporter.FileType()))

	buf := bytes.NewBuffer(nil)
	err = handler.exporter.Export(ctx, buf, res.Servers)
	if err != nil {
		return api.ExportServers500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}

	slog.Info("handler: export servers success", slog.Int("size", buf.Len()))

	contentDisposition := fmt.Sprintf(`attachment; filename="servers.%s"`, handler.exporter.FileType())
	return api.ExportServers200ApplicationvndOpenxmlformatsOfficedocumentSpreadsheetmlSheetResponse{
		Body: buf,
		Headers: api.ExportServers200ResponseHeaders{
			ContentDisposition: contentDisposition,
		},
	}, nil
}

// Import server
// (POST /servers/import)
func (handler *ServerRestHandler) ImportServer(ctx context.Context, request api.ImportServerRequestObject) (api.ImportServerResponseObject, error) {
	slog.Info("handler: import server")

	file, err := request.Body.NextPart()
	if err != nil {
		slog.Warn("failed to read file", slog.Any("err", err))
		return api.ImportServer400JSONResponse{
			BadRequestJSONResponse: BadRequest(err),
		}, nil
	}
	defer file.Close()

	slog.Info("handler: deserializing file", slog.String("file", file.FileName()))

	servers, err := handler.importer.Deserialize(ctx, file)
	if err != nil {
		switch {
		case errors.Is(err, apperr.ErrInvalidImportData):
			slog.Warn("invalid import data", slog.Any("err", err))
			return api.ImportServer400JSONResponse{
				BadRequestJSONResponse: BadRequest(err),
			}, nil
		default:
			slog.Error("failed to deserialize", slog.Any("err", err))
			return api.ImportServer500JSONResponse{
				InternalErrorJSONResponse: InternalError(err),
			}, nil
		}
	}

	slog.Info("handler: importing servers", slog.Int("total", len(servers)))

	res, err := handler.service.ImportServer(ctx, servers)
	if err != nil {
		slog.Error("failed to import servers", slog.Any("err", err))
		return api.ImportServer500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}

	slog.Info("handler: import servers success",
		slog.Int("total_success", res.SuccessCnt),
		slog.Int("total_failed", res.FailedCnt),
	)

	return api.ImportServer200JSONResponse{
		IdFailed:     res.Failed,
		IdSuccess:    res.Success,
		TotalFailed:  res.FailedCnt,
		TotalSuccess: res.SuccessCnt,
	}, nil
}
