package cache

import (
	"context"
	"time"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/model"
)

type DailyReportCacheInterface interface {
	Get(ctx context.Context, date time.Time) ([]model.ServerUptimeAgg, error)
	Set(ctx context.Context, date time.Time, data []model.ServerUptimeAgg) error
}
