package rdb

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const tokenBlocklistKeyFormat = "auth:blocklist:%s" // auth:blocklist:<token>

type TokenBlocklistRedis struct {
	client *redis.Client
}

func NewTokenBlocklistRedis(client *redis.Client) *TokenBlocklistRedis {
	return &TokenBlocklistRedis{client: client}
}

// Revoke thêm token vào Redis với TTL bằng thời gian còn lại đến khi token hết hạn.
// Redis tự xóa sau khi TTL hết — không cần cleanup thủ công.
func (b *TokenBlocklistRedis) Revoke(ctx context.Context, token string, expiry time.Time) error {
	ttl := time.Until(expiry)
	if ttl <= 0 {
		// Token đã hết hạn rồi, không cần lưu
		return nil
	}
	key := fmt.Sprintf(tokenBlocklistKeyFormat, token)
	if err := b.client.Set(ctx, key, 1, ttl).Err(); err != nil {
		return fmt.Errorf("tokenBlocklist.Revoke: %w", err)
	}
	return nil
}

// IsRevoked kiểm tra token có trong blocklist không.
func (b *TokenBlocklistRedis) IsRevoked(ctx context.Context, token string) (bool, error) {
	key := fmt.Sprintf(tokenBlocklistKeyFormat, token)
	err := b.client.Get(ctx, key).Err()
	if err == redis.Nil {
		return false, nil // không có trong blocklist
	}
	if err != nil {
		return false, fmt.Errorf("tokenBlocklist.IsRevoked: %w", err)
	}
	return true, nil
}
