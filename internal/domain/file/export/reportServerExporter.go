package export

import (
	"context"
	"io"

	"github.com/LeHuuHai/server-management/internal/model"
)

type ReportServerExporter interface {
	Export(ctx context.Context, writer io.Writer, data []model.ServerUptimeAgg) error
	FileName() string
	ContentType() string
}
