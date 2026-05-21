package main

import (
	"github.com/LeHuuHai/server-management/api"
	"github.com/LeHuuHai/server-management/config"
	"github.com/LeHuuHai/server-management/internal/handler"
	"github.com/LeHuuHai/server-management/internal/infra/app"
	xlsximport "github.com/LeHuuHai/server-management/internal/infra/file/deserialize"
	xlsxexport "github.com/LeHuuHai/server-management/internal/infra/file/export"
	pg "github.com/LeHuuHai/server-management/internal/infra/postgres"
	"github.com/LeHuuHai/server-management/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	app, err := app.New(cfg)
	if err != nil {
		panic(err)
	}

	// domain
	serverRepo := pg.NewServerRepository(app.DB)

	// service
	serverService := service.NewServerService(serverRepo)

	//handler
	serverHandler := handler.NewServerHandler(
		serverService,
		xlsxexport.NewServerXLSXExporter(),
		xlsximport.NewServerXLSXImporter(),
	)

	// router
	r := gin.New()
	api.RegisterHandlers(r, serverHandler)
	_ = r.Run(":8080")

}
