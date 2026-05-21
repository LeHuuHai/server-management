package deserialize

import (
	"context"
	"io"

	"github.com/LeHuuHai/server-management/internal/model"
)

type ServerDeserializer interface {
	Deserialize(ctx context.Context, reader io.Reader) ([]model.ServerImport, error)
}
