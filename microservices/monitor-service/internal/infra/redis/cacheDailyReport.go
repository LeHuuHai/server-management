package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/cache"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/model"
	"github.com/redis/go-redis/v9"
)

const dailyReportKeyFormat = "report:daily:%s"

type DailyReportRedisCache struct {
	client *redis.Client
}

func NewDailyReportRedisCache(client *redis.Client) cache.DailyReportCacheInterface {
	return &DailyReportRedisCache{client: client}
}

func dailyKey(date time.Time) string {
	return fmt.Sprintf(dailyReportKeyFormat, date.UTC().Format("2006-01-02"))
}

func (c *DailyReportRedisCache) Get(ctx context.Context, date time.Time) ([]model.ServerUptimeAgg, error) {
	val, err := c.client.Get(ctx, dailyKey(date)).Result()
	if err == redis.Nil {
		return nil, nil // cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("dailyReportCache.Get: %w", err)
	}

	var result []model.ServerUptimeAgg
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, fmt.Errorf("dailyReportCache.Get unmarshal: %w", err)
	}
	return result, nil
}

func (c *DailyReportRedisCache) Set(ctx context.Context, date time.Time, data []model.ServerUptimeAgg) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("dailyReportCache.Set marshal: %w", err)
	}
	// No TTL — daily data is immutable once the day ends
	if err := c.client.Set(ctx, dailyKey(date), b, 0).Err(); err != nil {
		return fmt.Errorf("dailyReportCache.Set: %w", err)
	}
	return nil
}
