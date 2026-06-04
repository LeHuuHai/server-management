package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/LeHuuHai/server-management/api"

	"github.com/LeHuuHai/server-management/internal/domain/file/deserialize"
	"github.com/LeHuuHai/server-management/internal/domain/file/export"
	apperr "github.com/LeHuuHai/server-management/internal/error"
	"github.com/LeHuuHai/server-management/internal/middleware"
	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/LeHuuHai/server-management/internal/service"
)

// impl StrictServerInterface
type ServerHandler struct {
	service       *service.ServerService
	reportService *service.ReportServerService
	exporter      export.ServerExporter
	deserialize   deserialize.ServerDeserializer
}

func NewServerHandler(s *service.ServerService, r *service.ReportServerService, e export.ServerExporter, d deserialize.ServerDeserializer) *ServerHandler {
	return &ServerHandler{
		service:       s,
		reportService: r,
		exporter:      e,
		deserialize:   d,
	}
}

// Get list servers
// (GET /servers)
func (handler *ServerHandler) GetListServers(ctx context.Context, request api.GetListServersRequestObject) (api.GetListServersResponseObject, error) {
	logger := middleware.LoggerFromContext(ctx)

	params := request.Params
	filter := model.ListServerFilter{
		From:      params.From,
		To:        params.To,
		SortField: model.ServerSortField(params.SortField),
		Desc:      params.Desc,
	}

	logger.Info("handler: get list servers", slog.Any("filter", filter))

	res, err := handler.service.ListServer(ctx, filter)
	if err != nil {
		if errors.Is(err, apperr.ErrInvalidSort) || errors.Is(err, apperr.ErrInvalidPagination) {
			logger.Warn("invalid request", slog.Any("err", err))
			return api.GetListServers400JSONResponse{
				BadRequestJSONResponse: BadRequest(err),
			}, nil
		}

		logger.Error("failed to get list servers", slog.Any("err", err))

		return api.GetListServers500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}

	logger.Info("handler: get list servers success", slog.Int("total", res.Total))

	items := make([]api.Server, len(res.Servers))
	for idx, s := range res.Servers {
		items[idx] = api.Server{
			ServerId:          s.ServerID,
			ServerName:        s.ServerName,
			Status:            api.ServerStatus(s.Status),
			Ipv4:              s.IPv4,
			CreatedAt:         &s.CreatedAt,
			MetadataUpdatedAt: &s.MetadataUpdatedAt,
			LastPingAt:        &s.LastPingAt,
		}
	}
	return api.GetListServers200JSONResponse{
		Items: &items,
		Total: &res.Total,
	}, nil
}

// Create server
// (POST /servers)
func (handler *ServerHandler) CreateServer(ctx context.Context, request api.CreateServerRequestObject) (api.CreateServerResponseObject, error) {
	logger := middleware.LoggerFromContext(ctx)

	server := model.Server{
		ServerID:   request.Body.ServerId,
		ServerName: request.Body.ServerName,
		IPv4:       request.Body.Ipv4,
	}

	logger.Info("handler: create servers", slog.Any("server", server))

	newServer, err := handler.service.CreateServer(ctx, &server)
	if err != nil {
		if errors.Is(err, apperr.ErrDuplicateServer) {
			logger.Warn("conflict request", slog.Any("err", err))
			return api.CreateServer409JSONResponse{
				ConflictJSONResponse: Conflict(err),
			}, nil
		}
		if errors.Is(err, apperr.ErrInvalidIP) {
			logger.Warn("invalid request", slog.Any("err", err))
			return api.CreateServer400JSONResponse{
				BadRequestJSONResponse: BadRequest(err),
			}, nil
		}

		logger.Error("failed to create servers", slog.Any("err", err))

		return api.CreateServer500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}

	logger.Info("handler: create servers success")

	return api.CreateServer201JSONResponse{
		ServerId:          newServer.ServerID,
		ServerName:        newServer.ServerName,
		Ipv4:              newServer.IPv4,
		Status:            api.ServerStatus(newServer.Status),
		CreatedAt:         &newServer.CreatedAt,
		LastPingAt:        &newServer.LastPingAt,
		MetadataUpdatedAt: &newServer.MetadataUpdatedAt,
	}, nil
}

// Export servers
// (GET /servers/export)
func (handler *ServerHandler) ExportServers(ctx context.Context, request api.ExportServersRequestObject) (api.ExportServersResponseObject, error) {
	logger := middleware.LoggerFromContext(ctx)

	params := request.Params
	filter := model.ListServerFilter{
		From:      params.From,
		To:        params.To,
		SortField: model.ServerSortField(params.SortField),
		Desc:      params.Desc,
	}

	logger.Info("handler: export servers", slog.Any("filter", filter))

	res, err := handler.service.ListServer(ctx, filter)
	if err != nil {
		if errors.Is(err, apperr.ErrInvalidSort) || errors.Is(err, apperr.ErrInvalidPagination) {
			logger.Warn("invalid request", slog.Any("err", err))
			return api.ExportServers400JSONResponse{
				BadRequestJSONResponse: BadRequest(err),
			}, nil
		}

		logger.Error("failed to list servers", slog.Any("err", err))

		return api.ExportServers500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}

	logger.Info("handler: exporting", slog.Int("total", res.Total), slog.String("file_type", handler.exporter.FileType()))

	buf := bytes.NewBuffer(nil)
	err = handler.exporter.Export(ctx, buf, res.Servers)
	if err != nil {
		return api.ExportServers500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}

	logger.Info("handler: export servers success", slog.Int("size", buf.Len()))

	contentDisposition := fmt.Sprintf(`attachment; filename="servers.%s"`, handler.exporter.FileType())
	return api.ExportServers200ApplicationoctetStreamResponse{
		Body: buf,
		Headers: api.ExportServers200ResponseHeaders{
			ContentDisposition: &contentDisposition,
		},
	}, nil
}

// Import server
// (POST /servers/import)
func (handler *ServerHandler) ImportServer(ctx context.Context, request api.ImportServerRequestObject) (api.ImportServerResponseObject, error) {
	logger := middleware.LoggerFromContext(ctx)

	logger.Info("handler: import server")

	file, err := request.Body.NextPart()
	if err != nil {
		logger.Warn("failed to read file", slog.Any("err", err))
		return api.ImportServer400JSONResponse{
			BadRequestJSONResponse: BadRequest(err),
		}, nil
	}
	defer file.Close()

	logger.Info("handler: deserializing file", slog.String("file", file.FileName()))

	servers, err := handler.deserialize.Deserialize(ctx, file)
	if err != nil {
		switch {
		case errors.Is(err, apperr.ErrInvalidImportData):
			logger.Warn("invalid import data", slog.Any("err", err))
			return api.ImportServer400JSONResponse{
				BadRequestJSONResponse: BadRequest(err),
			}, nil
		default:
			logger.Error("failed to deserialize", slog.Any("err", err))
			return api.ImportServer500JSONResponse{
				InternalErrorJSONResponse: InternalError(err),
			}, nil
		}
	}

	logger.Info("handler: importing servers", slog.Int("total", len(servers)))

	res, err := handler.service.ImportServer(ctx, servers)
	if err != nil {
		logger.Error("failed to import servers", slog.Any("err", err))
		return api.ImportServer500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}

	logger.Info("handler: import servers success",
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

// Delete server
// (DELETE /servers/{server_id})
func (handler *ServerHandler) DeleteServer(ctx context.Context, request api.DeleteServerRequestObject) (api.DeleteServerResponseObject, error) {
	logger := middleware.LoggerFromContext(ctx)

	logger.Info("handler: delete server", slog.String("server_id", request.ServerId))

	if err := handler.service.DeleteServer(ctx, request.ServerId); err != nil {
		if errors.Is(err, apperr.ErrRecordNotFound) {
			logger.Warn("server not found", slog.String("server_id", request.ServerId))
			return api.DeleteServer404JSONResponse{
				NotFoundJSONResponse: NotFound(err),
			}, nil
		}
		logger.Error("failed to delete server", slog.String("server_id", request.ServerId), slog.Any("err", err))
		return api.DeleteServer500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}

	logger.Info("handler: delete server success", slog.String("server_id", request.ServerId))

	return api.DeleteServer204Response{}, nil
}

// Update server
// (PATCH /servers/{server_id})
func (handler *ServerHandler) UpdateServer(ctx context.Context, request api.UpdateServerRequestObject) (api.UpdateServerResponseObject, error) {
	logger := middleware.LoggerFromContext(ctx)

	server := model.Server{
		ServerID:   request.ServerId,
		ServerName: *request.Body.ServerName,
		IPv4:       *request.Body.Ipv4,
	}

	logger.Info("handler: update server",
		slog.String("server_id", server.ServerID),
		slog.String("server_name", server.ServerName),
		slog.String("ipv4", server.IPv4),
	)

	s, err := handler.service.UpdateServer(ctx, &server)
	if err != nil {
		if errors.Is(err, apperr.ErrRecordNotFound) {
			logger.Warn("server not found", slog.String("server_id", server.ServerID))
			return api.UpdateServer404JSONResponse{
				NotFoundJSONResponse: NotFound(err),
			}, nil
		}
		logger.Error("failed to update server", slog.String("server_id", server.ServerID), slog.Any("err", err))
		return api.UpdateServer500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}

	logger.Info("handler: update server success", slog.String("server_id", s.ServerID))

	return api.UpdateServer200JSONResponse{
		ServerId:          s.ServerID,
		ServerName:        s.ServerName,
		Status:            api.ServerStatus(s.Status),
		Ipv4:              s.IPv4,
		CreatedAt:         &s.CreatedAt,
		MetadataUpdatedAt: &s.MetadataUpdatedAt,
		LastPingAt:        &s.LastPingAt,
	}, nil
}

// Generate server report
// (POST /servers/report)
func (handler *ServerHandler) GenerateServerReport(ctx context.Context, request api.GenerateServerReportRequestObject) (api.GenerateServerReportResponseObject, error) {
	logger := middleware.LoggerFromContext(ctx)

	receivers := make([]string, len(request.Body.Receivers))
	for i, r := range request.Body.Receivers {
		receivers[i] = string(r)
	}

	req := model.GenServerReportRequest{
		From:      request.Body.From,
		To:        request.Body.To,
		Receivers: receivers,
	}

	logger.Info("handler: generate server report",
		slog.Any("from", req.From),
		slog.Any("to", req.To),
		slog.Any("receivers", req.Receivers),
	)

	err := handler.reportService.ReportServer(ctx, req)
	if err != nil {
		if errors.Is(err, apperr.ErrInvalidTimeRange) || errors.Is(err, apperr.ErrInvalidEmail) {
			logger.Warn("invalid request", slog.Any("err", err))
			return api.GenerateServerReport400JSONResponse{
				BadRequestJSONResponse: BadRequest(err),
			}, nil
		}
		logger.Error("failed to generate server report", slog.Any("err", err))
		return api.GenerateServerReport500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}

	logger.Info("handler: generate server report accepted",
		slog.Any("receivers", req.Receivers),
	)

	return api.GenerateServerReport202Response{}, nil
}

// Download report file
// (GET /report/{filename})
func (handler *ServerHandler) GetReportFile(ctx context.Context, request api.GetReportFileRequestObject) (api.GetReportFileResponseObject, error) {
	logger := middleware.LoggerFromContext(ctx)

	filename := filepath.Base(request.Filename)
	path := filepath.Join("tmp", filename)

	logger.Info("handler: get report file", slog.String("filename", filename))

	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Warn("report file not found", slog.String("path", path))
			return api.GetReportFile404JSONResponse{
				NotFoundJSONResponse: NotFound(err),
			}, nil
		}
		logger.Error("failed to stat report file", slog.String("path", path), slog.Any("err", err))
		return api.GetReportFile500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}

	file, err := os.Open(path)
	if err != nil {
		logger.Error("failed to open report file", slog.String("path", path), slog.Any("err", err))
		return api.GetReportFile500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}

	logger.Info("handler: get report file success",
		slog.String("filename", filename),
		slog.Int64("size", info.Size()),
	)

	return api.GetReportFile200ApplicationoctetStreamResponse{
		Body:          file,
		ContentLength: info.Size(),
	}, nil
}
