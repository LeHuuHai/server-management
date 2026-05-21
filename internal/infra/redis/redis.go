package rdb

import (
	"context"
	"log"

	"github.com/LeHuuHai/server-management/config"
	"github.com/redis/go-redis/v9"
)

func Connect(config *config.MasterConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Redis.RedisURL,
		Password: config.Redis.RedisPassword,
		DB:       config.Redis.RedisDB,
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("connect Redis error: %s", err.Error())
		return nil, err
	}
	return rdb, nil
}
