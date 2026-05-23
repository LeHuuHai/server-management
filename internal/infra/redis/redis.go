package rdb

import (
	"context"
	"log"

	commonconfig "github.com/LeHuuHai/server-management/config/common"
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
		log.Printf("connect Redis error: %s", err.Error())
		return nil, err
	}
	return rdb, nil
}
