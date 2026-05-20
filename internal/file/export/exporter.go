package export

import (
	"context"
	"io"
)

type Exporter interface {
	Export(ctx context.Context, writer io.Writer, data any) error
	FileName() string
	ContentType() string
}
