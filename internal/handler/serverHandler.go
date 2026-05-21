package handler

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/LeHuuHai/server-management/api"

	apperr "github.com/LeHuuHai/server-management/internal/error"
	"github.com/LeHuuHai/server-management/internal/file/deserialize"
	"github.com/LeHuuHai/server-management/internal/file/export"
	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/LeHuuHai/server-management/internal/service"
	"github.com/gin-gonic/gin"
)

// impl ServerInterface
type ServerHandler struct {
	service     *service.ServerService
	exporter    export.ServerExporter
	deserialize deserialize.ServerDeserializer
}

func NewServerHandler(s *service.ServerService, e export.ServerExporter, d deserialize.ServerDeserializer) *ServerHandler {
	return &ServerHandler{
		service:     s,
		exporter:    e,
		deserialize: d,
	}
}

// Get list servers
// (GET /servers)
func (handler *ServerHandler) GetListServers(c *gin.Context, params api.GetListServersParams) {
	filter := model.ListServerFilter{
		From:      params.From,
		To:        params.To,
		SortField: model.ServerSortField(params.SortField),
		Desc:      params.Desc,
	}
	res, err := handler.service.ListServer(c.Request.Context(), filter)
	if err != nil {
		if errors.Is(err, apperr.ErrInvalidSort) || errors.Is(err, apperr.ErrInvalidPagination) {
			c.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}
	items := make([]api.Server, len(res.Servers))
	for idx, it := range res.Servers {
		items[idx] = ServerModelToServerAPI(it)
	}
	c.JSON(
		200,
		api.GetListServersResponse{
			Items: &items,
			Total: &res.Total,
		},
	)
}

// Create server
// (POST /servers)
func (handler *ServerHandler) CreateServer(c *gin.Context) {
	var dto model.ServerRequestDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}
	newServer := ServerDTOToServerModel(dto)
	if err := handler.service.CreateServer(c.Request.Context(), &newServer); err != nil {
		if errors.Is(err, apperr.ErrDuplicateServer) {
			c.JSON(409, gin.H{
				"message": err.Error(),
			})
			return
		}
		if errors.Is(err, apperr.ErrInvalidIP) {
			c.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(201, gin.H{
		"message": "success",
	})
}

// Export servers
// (GET /servers/export)
func (handler *ServerHandler) ExportServers(c *gin.Context, params api.ExportServersParams) {
	filter := model.ListServerFilter{
		From:      params.From,
		To:        params.To,
		SortField: model.ServerSortField(params.SortField),
		Desc:      params.Desc,
	}
	res, err := handler.service.ListServer(c.Request.Context(), filter)
	if err != nil {
		if errors.Is(err, apperr.ErrInvalidSort) || errors.Is(err, apperr.ErrInvalidPagination) {
			c.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}
	buf := bytes.NewBuffer(nil)
	err = handler.exporter.Export(c.Request.Context(), buf, res.Servers)
	if err != nil {
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.Header(
		"Content-Disposition",
		fmt.Sprintf(
			`attachment; filename="%s"`,
			handler.exporter.FileName(),
		),
	)

	c.Data(
		200,
		handler.exporter.ContentType(),
		buf.Bytes(),
	)
}

// Import server
// (POST /servers/import)
func (handler *ServerHandler) ImportServer(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{
			"message": "file is required",
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}
	defer file.Close()

	servers, err := handler.deserialize.Deserialize(c.Request.Context(), file)
	if err != nil {
		switch {
		case errors.Is(err, apperr.ErrInvalidImportData):
			c.JSON(400, gin.H{
				"message": err.Error(),
			})

		default:
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
		}
		return
	}

	res, err := handler.service.ImportServer(c.Request.Context(), servers)
	if err != nil {
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200,
		api.ImportServerResponse{
			IdFailed:     res.Failed,
			IdSuccess:    res.Success,
			TotalFailed:  res.FailedCnt,
			TotalSuccess: res.SuccessCnt,
		},
	)
}

// Delete server
// (DELETE /servers/{server_id})
func (handler *ServerHandler) DeleteServer(c *gin.Context, serverId string) {
	if err := handler.service.DeleteServer(c.Request.Context(), serverId); err != nil {
		if errors.Is(err, apperr.ErrRecordNotFound) {
			c.JSON(404, gin.H{
				"message": err.Error(),
			})
			return
		}
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.Status(204)
}

// Update server
// (PATCH /servers/{server_id})
func (handler *ServerHandler) UpdateServer(c *gin.Context, serverId string) {
	var dto model.ServerRequestDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}
	server := ServerDTOToServerModel(dto)
	server.ServerID = serverId
	s, err := handler.service.UpdateServer(c.Request.Context(), &server)
	if err != nil {
		if errors.Is(err, apperr.ErrRecordNotFound) {
			c.JSON(404, gin.H{
				"message": err.Error(),
			})
			return
		}
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"Server": ServerModelToServerAPI(*s),
	})
}
