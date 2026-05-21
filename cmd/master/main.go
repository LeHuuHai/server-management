package main

import (
	"github.com/LeHuuHai/server-management/api"
	"github.com/LeHuuHai/server-management/internal/domain/repo"
	"github.com/LeHuuHai/server-management/internal/file/deserialize"
	"github.com/LeHuuHai/server-management/internal/file/export"
	"github.com/LeHuuHai/server-management/internal/handler"
	"github.com/LeHuuHai/server-management/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {

	// domain
	serverRepo := repo.NewServerRepository(db)

	// service
	serverService := service.NewServerService(serverRepo)

	//handler
	serverHandler := handler.NewServerHandler(
		serverService,
		&export.ServerXLSXExporter{},
		&deserialize.ServerXLSXImporter{},
	)

	// router
	r := gin.New()
	api.RegisterHandlers(r, serverHandler)
	_ = r.Run(":8080")

}
