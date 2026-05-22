package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/elastic/go-elasticsearch/v8"
)

type Aggregator struct {
	client *elasticsearch.Client
}

type aggregationResponse struct {
	Aggregations struct {
		ByServer struct {
			Buckets []struct {
				Key string `json:"key"`

				UptimeRatio struct {
					Value float64 `json:"value"`
				} `json:"uptime_ratio"`

				FirstTimestamp struct {
					ValueAsString string `json:"value_as_string"`
				} `json:"first_timestamp"`

				LastTimestamp struct {
					ValueAsString string `json:"value_as_string"`
				} `json:"last_timestamp"`
			} `json:"buckets"`
		} `json:"by_server"`
	} `json:"aggregations"`
}

func (aggregator *Aggregator) Aggregation(ctx context.Context, from time.Time, to time.Time) ([]model.ServerUptimeAgg, error) {
	query := map[string]any{
		"size": 0,
		"query": map[string]any{
			"range": map[string]any{
				"timestamp": map[string]any{
					"gte": from.Format(time.RFC3339),
					"lt":  to.Format(time.RFC3339),
				},
			},
		},
		"aggs": map[string]any{
			"by_server": map[string]any{
				"terms": map[string]any{
					"field": "server_id.keyword",
				},
				"aggs": map[string]any{
					"by_status_on": map[string]any{
						"filter": map[string]any{
							"term": map[string]any{
								"status.keyword": "on",
							},
						},
					},
					"status_count": map[string]any{
						"value_count": map[string]any{
							"field": "status.keyword",
						},
					},
					"uptime_ratio": map[string]any{
						"bucket_script": map[string]any{
							"buckets_path": map[string]any{
								"on":    "by_status_on._count",
								"total": "status_count",
							},
							"script": "params.on / params.total",
						},
					},
					"first_timestamp": map[string]any{
						"min": map[string]any{
							"field":  "timestamp",
							"format": "strict_date_time",
						},
					},
					"last_timestamp": map[string]any{
						"max": map[string]any{
							"field":  "timestamp",
							"format": "strict_date_time",
						},
					},
				},
			},
		},
	}

	body, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	res, err := aggregator.client.Search(
		aggregator.client.Search.WithIndex("pings"),
		aggregator.client.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("Aggregation error: %s", body)
	}

	var parsed aggregationResponse
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	results := make([]model.ServerUptimeAgg, 0)

	for _, b := range parsed.Aggregations.ByServer.Buckets {
		s, err := time.Parse(time.RFC3339Nano, b.FirstTimestamp.ValueAsString)
		if err != nil {
			s = time.Time{}
		}
		l, err := time.Parse(time.RFC3339Nano, b.LastTimestamp.ValueAsString)
		if err != nil {
			l = time.Time{}
		}
		results = append(results, model.ServerUptimeAgg{
			ServerID:    b.Key,
			UptimeRatio: b.UptimeRatio.Value,
			StartPingAt: s,
			LastPingAt:  l,
		})
	}

	return results, nil
}
