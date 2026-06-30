package deserialize

import (
	"context"
	"io"

	"github.com/LeHuuHai/server-management/microservices/server-service/internal/model"
)

type ServerDeserializer interface {
	Deserialize(ctx context.Context, reader io.Reader) ([]model.ServerAddress, error)
}
