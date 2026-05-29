package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/LeHuuHai/server-management/api"

	"github.com/LeHuuHai/server-management/internal/domain/file/deserialize"
	"github.com/LeHuuHai/server-management/internal/domain/file/export"
	apperr "github.com/LeHuuHai/server-management/internal/error"
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
	params := request.Params
	filter := model.ListServerFilter{
		From:      params.From,
		To:        params.To,
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
	server := model.Server{
		ServerID:   request.Body.ServerId,
		ServerName: request.Body.ServerName,
		IPv4:       request.Body.Ipv4,
	}
	if err := handler.service.CreateServer(ctx, &server); err != nil {
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
	return api.CreateServer201JSONResponse{}, nil
}

// Export servers
// (GET /servers/export)
func (handler *ServerHandler) ExportServers(ctx context.Context, request api.ExportServersRequestObject) (api.ExportServersResponseObject, error) {
	params := request.Params
	filter := model.ListServerFilter{
		From:      params.From,
		To:        params.To,
		SortField: model.ServerSortField(params.SortField),
		Desc:      params.Desc,
	}
	res, err := handler.service.ListServer(ctx, filter)
	if err != nil {
		if errors.Is(err, apperr.ErrInvalidSort) || errors.Is(err, apperr.ErrInvalidPagination) {
			return api.ExportServers400JSONResponse{
				BadRequestJSONResponse: BadRequest(err),
			}, nil
		}
		return api.ExportServers500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}
	buf := bytes.NewBuffer(nil)
	err = handler.exporter.Export(ctx, buf, res.Servers)
	if err != nil {
		return api.ExportServers500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}
	return api.ExportServers200ApplicationoctetStreamResponse{
		Body: buf,
		Headers: api.ExportServers200ResponseHeaders{
			ContentDisposition: fmt.Sprintf(`attachment; filename="servers.%s"`, handler.exporter.FileType()),
		},
	}, nil
}

// Import server
// (POST /servers/import)
func (handler *ServerHandler) ImportServer(ctx context.Context, request api.ImportServerRequestObject) (api.ImportServerResponseObject, error) {
	file, err := request.Body.NextPart()
	if err != nil {
		return api.ImportServer400JSONResponse{
			BadRequestJSONResponse: BadRequest(err),
		}, nil
	}
	defer file.Close()

	servers, err := handler.deserialize.Deserialize(ctx, file)
	if err != nil {
		switch {
		case errors.Is(err, apperr.ErrInvalidImportData):
			return api.ImportServer400JSONResponse{
				BadRequestJSONResponse: BadRequest(err),
			}, nil
		default:
			return api.ImportServer500JSONResponse{
				InternalErrorJSONResponse: InternalError(err),
			}, nil
		}
	}

	res, err := handler.service.ImportServer(ctx, servers)
	if err != nil {
		return api.ImportServer500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}
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

// Update server
// (PATCH /servers/{server_id})
func (handler *ServerHandler) UpdateServer(ctx context.Context, request api.UpdateServerRequestObject) (api.UpdateServerResponseObject, error) {
	server := model.Server{
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
	receivers := make([]string, len(request.Body.Receivers))
	for i, r := range request.Body.Receivers {
		receivers[i] = string(r)
	}
	req := model.GenServerReportRequest{
		From:      request.Body.From,
		To:        request.Body.To,
		Receivers: receivers,
	}
	err := handler.reportService.ReportServer(ctx, req)
	if err != nil {
		if errors.Is(err, apperr.ErrInvalidTimeRange) || errors.Is(err, apperr.ErrInvalidEmail) {
			return api.GenerateServerReport400JSONResponse{
				BadRequestJSONResponse: BadRequest(err),
			}, nil
		}
		return api.GenerateServerReport500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}
	return api.GenerateServerReport202Response{}, nil
}

// Download report file
// (GET /report/{filename})
func (handler *ServerHandler) GetReportFile(ctx context.Context, request api.GetReportFileRequestObject) (api.GetReportFileResponseObject, error) {
	filename := filepath.Base(request.Filename)
	path := filepath.Join("tmp", filename)

	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return api.GetReportFile404JSONResponse{
				NotFoundJSONResponse: NotFound(err),
			}, nil
		}

		return api.GetReportFile500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return api.GetReportFile500JSONResponse{
			InternalErrorJSONResponse: InternalError(err),
		}, nil
	}

	return api.GetReportFile200ApplicationoctetStreamResponse{
		Body:          file,
		ContentLength: info.Size(),
	}, nil
}
