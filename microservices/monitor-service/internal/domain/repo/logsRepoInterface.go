package repo

import (
	"context"
)

type LogsRepositoryInterface[T any] interface {
	WriteBatch(ctx context.Context, logs []T) error
}
