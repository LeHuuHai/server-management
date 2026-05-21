package export

import (
	"context"
	"io"

	"github.com/LeHuuHai/server-management/internal/model"
)

type ServerExporter interface {
	Export(ctx context.Context, writer io.Writer, data []model.Server) error
	FileName() string
	ContentType() string
}
