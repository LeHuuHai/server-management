package export

import (
	"context"
	"io"

	"github.com/LeHuuHai/server-management/microservices/server-service/internal/model"
)

type ServerExporter interface {
	Export(ctx context.Context, writer io.Writer, data []model.ServerProfile) error
	FileType() string
	ContentType() string
}
