package cache

import (
	"context"
	"fmt"
	"log/slog"

	pkgconfig "github.com/LeHuuHai/server-management/microservices/pkg/config"
	"github.com/redis/go-redis/v9"
)

func Connect(config *pkgconfig.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.URL,
		Password: config.Password,
		DB:       config.DB,
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("connect redis error: %v", err)
	}
	slog.Info("Redis connected", "url", config.URL, "db", config.DB)
	return rdb, nil
}
