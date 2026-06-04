package aggregator

import (
	"context"
	"time"

	"github.com/LeHuuHai/server-management/internal/model"
)

type ReportAggregator interface {
	Aggregation(ctx context.Context, from time.Time, to time.Time) ([]model.ServerUptimeAgg, error)
}
