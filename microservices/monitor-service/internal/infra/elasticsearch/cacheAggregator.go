package elasticsearch

import (
	"context"
	"time"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/aggregator"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/cache"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/model"
)

// CachedAggregator bọc Aggregator gốc, thêm daily Redis cache.
//
// Chiến lược khi nhận range [from, to):
//
//	[from]-----[dayStart]----------[dayEnd]-----[to]
//	  ↑ phần lẻ đầu (query ES)       ↑ phần lẻ cuối (query ES)
//	              ↑ các ngày hoàn chỉnh đã kết thúc (dùng cache)
//
// "Ngày hoàn chỉnh đã kết thúc" = ngày mà cả [00:00, 00:00 hôm sau) đều < now.
type CachedAggregator struct {
	aggregator aggregator.ReportAggregator
	cache      cache.DailyReportCacheInterface
}

func NewCachedAggregator(agg aggregator.ReportAggregator, c cache.DailyReportCacheInterface) aggregator.ReportAggregator {
	return &CachedAggregator{
		aggregator: agg,
		cache:      c,
	}
}

// Aggregation là entry point, thay thế Aggregator.Aggregation trong ReportServerService.
func (c *CachedAggregator) Aggregation(ctx context.Context, from time.Time, to time.Time) ([]model.ServerUptimeAgg, error) {
	from = from.UTC()
	to = to.UTC()
	now := time.Now().UTC()

	// Mốc đầu ngày của today (00:00 UTC) — mọi ngày < todayStart là đã kết thúc
	todayStart := truncateToDay(now)

	// Tính các daily chunk hoàn chỉnh nằm trong [from, to) và đã kết thúc
	// dayStart = ngày đầu tiên có 00:00 >= from
	// dayEnd   = ngày cuối có 00:00+24h <= to VÀ <= todayStart
	chunkStart := truncateToDay(from)
	if chunkStart.Before(from) {
		chunkStart = chunkStart.Add(24 * time.Hour)
	}
	chunkEnd := truncateToDay(to) // 00:00 của ngày chứa `to`
	if chunkEnd.After(todayStart) {
		chunkEnd = todayStart
	}

	// Thu thập tất cả kết quả
	allResults := make([]model.ServerUptimeAgg, 0)

	// 1. Phần lẻ đầu: [from, chunkStart) — nếu có và chưa cache được
	if from.Before(chunkStart) {
		partial, err := c.aggregator.Aggregation(ctx, from, chunkStart)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, partial...)
	}

	// 2. Các ngày hoàn chỉnh: [chunkStart, chunkEnd) theo từng ngày
	for day := chunkStart; day.Before(chunkEnd); day = day.Add(24 * time.Hour) {
		dayResult, err := c.getOrFetchDay(ctx, day)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, dayResult...)
	}

	// 3. Phần lẻ cuối: [chunkEnd, to) — phần còn lại chưa đủ một ngày
	//    (hoặc ngày hiện tại đang chạy, không cache)
	if chunkEnd.Before(to) {
		partial, err := c.aggregator.Aggregation(ctx, chunkEnd, to)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, partial...)
	}

	return mergeByServerID(allResults), nil
}

// getOrFetchDay lấy cache của một ngày, nếu miss thì query ES rồi set cache.
func (c *CachedAggregator) getOrFetchDay(ctx context.Context, dayStart time.Time) ([]model.ServerUptimeAgg, error) {
	cached, err := c.cache.Get(ctx, dayStart)
	if err != nil {
		return nil, err
	}
	if cached != nil {
		return cached, nil // cache hit
	}

	// cache miss → query ES
	dayEnd := dayStart.Add(24 * time.Hour)
	fresh, err := c.aggregator.Aggregation(ctx, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}

	// set cache (fire-and-forget lỗi không làm fail request)
	_ = c.cache.Set(ctx, dayStart, fresh)

	return fresh, nil
}

// mergeByServerID gộp các ServerUptimeAgg có cùng ServerID từ nhiều chunk lại.
// Uptime ratio được tính lại theo weighted average dựa trên DocCount.
func mergeByServerID(items []model.ServerUptimeAgg) []model.ServerUptimeAgg {
	type accumulator struct {
		totalDocs   int64
		weightedSum float64 // sum(uptimeRatio * docCount)
		startPingAt time.Time
		lastPingAt  time.Time
	}

	acc := make(map[string]*accumulator)

	for _, item := range items {
		a, exists := acc[item.ServerID]
		if !exists {
			acc[item.ServerID] = &accumulator{
				totalDocs:   item.DocCount,
				weightedSum: item.UptimeRatio * float64(item.DocCount),
				startPingAt: item.StartPingAt,
				lastPingAt:  item.LastPingAt,
			}
			continue
		}

		a.weightedSum += item.UptimeRatio * float64(item.DocCount)
		a.totalDocs += item.DocCount

		if item.StartPingAt.Before(a.startPingAt) {
			a.startPingAt = item.StartPingAt
		}
		if item.LastPingAt.After(a.lastPingAt) {
			a.lastPingAt = item.LastPingAt
		}
	}

	result := make([]model.ServerUptimeAgg, 0, len(acc))
	for serverID, a := range acc {
		ratio := 0.0
		if a.totalDocs > 0 {
			ratio = a.weightedSum / float64(a.totalDocs)
		}
		result = append(result, model.ServerUptimeAgg{
			ServerID:    serverID,
			UptimeRatio: ratio,
			StartPingAt: a.startPingAt,
			LastPingAt:  a.lastPingAt,
			DocCount:    a.totalDocs,
		})
	}
	return result
}

// truncateToDay cắt time về 00:00:00 UTC của ngày đó.
func truncateToDay(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
