package rdb_test

import (
	"context"
	"testing"
	"time"

	rdb "github.com/LeHuuHai/server-management/internal/infra/redis"
	"github.com/LeHuuHai/server-management/internal/model"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	rediscontainer "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ---------------------------------------------------------------------------
// Suite setup
// ---------------------------------------------------------------------------

type RedisIntegrationSuite struct {
	suite.Suite
	container testcontainers.Container
	client    *goredis.Client
}

func (s *RedisIntegrationSuite) SetupSuite() {
	ctx := context.Background()

	container, err := rediscontainer.RunContainer(ctx,
		testcontainers.WithImage("redis:7-alpine"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(s.T(), err)
	s.container = container

	endpoint, err := container.Endpoint(ctx, "")
	require.NoError(s.T(), err)

	s.client = goredis.NewClient(&goredis.Options{Addr: endpoint})
}

func (s *RedisIntegrationSuite) TearDownSuite() {
	s.client.Close()
	s.container.Terminate(context.Background())
}

func (s *RedisIntegrationSuite) SetupTest() {
	s.client.FlushAll(context.Background())
}

func TestRedisIntegrationSuite(t *testing.T) {
	suite.Run(t, new(RedisIntegrationSuite))
}

// ---------------------------------------------------------------------------
// TokenBlocklist tests
// ---------------------------------------------------------------------------

func (s *RedisIntegrationSuite) TestTokenBlocklist_Revoke_AndIsRevoked() {
	blocklist := rdb.NewTokenBlocklistRedis(s.client)
	ctx := context.Background()

	token := "test-token"
	expiry := time.Now().Add(time.Hour)

	err := blocklist.Revoke(ctx, token, expiry)
	assert.NoError(s.T(), err)

	revoked, err := blocklist.IsRevoked(ctx, token)
	assert.NoError(s.T(), err)
	assert.True(s.T(), revoked)
}

func (s *RedisIntegrationSuite) TestTokenBlocklist_IsRevoked_NotRevoked() {
	blocklist := rdb.NewTokenBlocklistRedis(s.client)

	revoked, err := blocklist.IsRevoked(context.Background(), "unknown-token")
	assert.NoError(s.T(), err)
	assert.False(s.T(), revoked)
}

func (s *RedisIntegrationSuite) TestTokenBlocklist_Revoke_ExpiredToken_NotStored() {
	blocklist := rdb.NewTokenBlocklistRedis(s.client)
	ctx := context.Background()

	// expiry đã qua → không lưu
	err := blocklist.Revoke(ctx, "expired-token", time.Now().Add(-time.Hour))
	assert.NoError(s.T(), err)

	revoked, err := blocklist.IsRevoked(ctx, "expired-token")
	assert.NoError(s.T(), err)
	assert.False(s.T(), revoked)
}

// ---------------------------------------------------------------------------
// DailyReportCache tests
// ---------------------------------------------------------------------------

func (s *RedisIntegrationSuite) TestDailyReportCache_SetAndGet_Success() {
	cache := rdb.NewDailyReportRedisCache(s.client)
	ctx := context.Background()
	date := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)

	data := []model.ServerUptimeAgg{
		{ServerID: "s1", UptimeRatio: 0.99, DocCount: 100},
		{ServerID: "s2", UptimeRatio: 0.75, DocCount: 50},
	}

	err := cache.Set(ctx, date, data)
	assert.NoError(s.T(), err)

	result, err := cache.Get(ctx, date)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), result, 2)
	assert.Equal(s.T(), "s1", result[0].ServerID)
	assert.InDelta(s.T(), 0.99, result[0].UptimeRatio, 0.001)
}

func (s *RedisIntegrationSuite) TestDailyReportCache_Get_CacheMiss_ReturnsNil() {
	cache := rdb.NewDailyReportRedisCache(s.client)

	result, err := cache.Get(context.Background(), time.Now())
	assert.NoError(s.T(), err)
	assert.Nil(s.T(), result)
}

func (s *RedisIntegrationSuite) TestDailyReportCache_Set_SameDateOverwrites() {
	cache := rdb.NewDailyReportRedisCache(s.client)
	ctx := context.Background()
	date := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)

	err := cache.Set(ctx, date, []model.ServerUptimeAgg{{ServerID: "s1", UptimeRatio: 0.99}})
	require.NoError(s.T(), err)

	// overwrite
	err = cache.Set(ctx, date, []model.ServerUptimeAgg{{ServerID: "s2", UptimeRatio: 0.50}})
	require.NoError(s.T(), err)

	result, err := cache.Get(ctx, date)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), result, 1)
	assert.Equal(s.T(), "s2", result[0].ServerID)
}

func (s *RedisIntegrationSuite) TestDailyReportCache_DifferentDates_Independent() {
	cache := rdb.NewDailyReportRedisCache(s.client)
	ctx := context.Background()

	date1 := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)

	err := cache.Set(ctx, date1, []model.ServerUptimeAgg{{ServerID: "s1"}})
	require.NoError(s.T(), err)
	err = cache.Set(ctx, date2, []model.ServerUptimeAgg{{ServerID: "s2"}})
	require.NoError(s.T(), err)

	res1, _ := cache.Get(ctx, date1)
	res2, _ := cache.Get(ctx, date2)

	assert.Equal(s.T(), "s1", res1[0].ServerID)
	assert.Equal(s.T(), "s2", res2[0].ServerID)
}
