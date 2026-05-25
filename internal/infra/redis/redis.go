package rdb

import (
	"context"
	"fmt"

	commonconfig "github.com/LeHuuHai/server-management/config/common"
	apperr "github.com/LeHuuHai/server-management/internal/error"
	"github.com/redis/go-redis/v9"
)

func Connect(config *commonconfig.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.URL,
		Password: config.Password,
		DB:       config.DB,
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrConnectRedis, err)
	}
	return rdb, nil
}
