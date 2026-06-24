package rdb

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/LeHuuHai/server-management/microservices/pkg/apperr"
	pkgconfig "github.com/LeHuuHai/server-management/microservices/pkg/config"
	"github.com/redis/go-redis/v9"
)

func Connect(config *pkgconfig.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.URL,
		Password: config.Password,
		DB:       config.DB,
	})

	err := rdb.Ping(context.Background()).Err()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrConnectRedis, err)
	}

	slog.Info("Redis connected", "url", config.URL, "db", config.DB)
	return rdb, nil
}
