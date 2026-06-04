package cache

import (
	"context"
	"time"
)

type TokenBlocklist interface {
	Revoke(ctx context.Context, token string, expiry time.Time) error

	IsRevoked(ctx context.Context, token string) (bool, error)
}
