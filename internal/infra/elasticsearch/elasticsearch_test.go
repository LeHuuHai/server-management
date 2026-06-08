package es_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	es "github.com/LeHuuHai/server-management/internal/infra/elasticsearch"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	tces "github.com/testcontainers/testcontainers-go/modules/elasticsearch"
	"github.com/testcontainers/testcontainers-go/wait"
)

const testIndex = "heartbeats_test"

type ESIntegrationSuite struct {
	suite.Suite
	container *tces.ElasticsearchContainer
	client    *elasticsearch.Client
	agg       *es.Aggregator
}

func (s *ESIntegrationSuite) SetupSuite() {
	ctx := context.Background()

	container, err := tces.RunContainer(ctx,
		testcontainers.WithImage("docker.elastic.co/elasticsearch/elasticsearch:8.13.4"),
		testcontainers.WithEnv(map[string]string{
			"discovery.type":         "single-node",
			"xpack.security.enabled": "false",
			"ES_JAVA_OPTS":           "-Xms512m -Xmx512m",
		}),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/_cluster/health").
				WithPort("9200/tcp").
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(s.T(), err)
	s.container = container

	host, err := container.Host(ctx)
	require.NoError(s.T(), err)
	esPort, err := container.MappedPort(ctx, "9200")
	require.NoError(s.T(), err)

	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{fmt.Sprintf("http://%s:%s", host, esPort.Port())},
	})
	require.NoError(s.T(), err)

	s.client = client
	s.agg = es.NewESAggregator(client, testIndex)

	// tạo index với mapping đúng — server_id và status phải là keyword để aggregation được
	mapping := `{
		"mappings": {
			"properties": {
				"server_id": { "type": "keyword" },
				"status":    { "type": "keyword" },
				"timestamp": { "type": "date" }
			}
		}
	}`
	s.client.Indices.Create(testIndex, s.client.Indices.Create.WithBody(strings.NewReader(mapping)))
}

func (s *ESIntegrationSuite) TearDownSuite() {
	s.client.Indices.Delete([]string{testIndex})
	s.container.Terminate(context.Background())
}

func (s *ESIntegrationSuite) SetupTest() {
	// xóa hết document rồi force refresh
	s.client.DeleteByQuery(
		[]string{testIndex},
		strings.NewReader(`{"query":{"match_all":{}}}`),
		s.client.DeleteByQuery.WithRefresh(true),
	)
	s.client.Indices.Refresh(s.client.Indices.Refresh.WithIndex(testIndex))
}

func TestESIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ESIntegrationSuite))
}

func (s *ESIntegrationSuite) indexHeartbeat(serverID string, status string, ts time.Time) {
	doc := map[string]any{
		"server_id": serverID,
		"status":    status,
		"timestamp": ts.Format(time.RFC3339),
	}
	body, _ := json.Marshal(doc)
	s.client.Index(testIndex, strings.NewReader(string(body)))
}

// ---------------------------------------------------------------------------
// Aggregation tests
// ---------------------------------------------------------------------------

func (s *ESIntegrationSuite) TestAggregation_NoData_ReturnsEmpty() {
	from := time.Now().Add(-time.Hour)
	to := time.Now()

	result, err := s.agg.Aggregation(context.Background(), from, to)
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), result)
}

func (s *ESIntegrationSuite) TestAggregation_Success() {
	now := time.Now()

	// s1: 3 on, 1 off → uptime 0.75
	for i := 0; i < 3; i++ {
		s.indexHeartbeat("s1", "on", now.Add(-time.Duration(i)*time.Minute))
	}
	s.indexHeartbeat("s1", "off", now.Add(-4*time.Minute))

	// s2: 2 on → uptime 1.0
	for i := 0; i < 2; i++ {
		s.indexHeartbeat("s2", "on", now.Add(-time.Duration(i)*time.Minute))
	}

	// đợi ES index xong
	s.client.Indices.Refresh()
	time.Sleep(500 * time.Millisecond)

	from := now.Add(-time.Hour)
	to := now.Add(time.Minute)

	result, err := s.agg.Aggregation(context.Background(), from, to)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), result, 2)

	byID := make(map[string]float64)
	for _, r := range result {
		byID[r.ServerID] = r.UptimeRatio
	}
	assert.InDelta(s.T(), 0.75, byID["s1"], 0.01)
	assert.InDelta(s.T(), 1.0, byID["s2"], 0.01)
}

func (s *ESIntegrationSuite) TestAggregation_OutsideTimeRange_ReturnsEmpty() {
	now := time.Now()
	s.indexHeartbeat("s1", "on", now.Add(-2*time.Hour))
	s.client.Indices.Refresh()
	time.Sleep(500 * time.Millisecond)

	// query range không chứa document
	from := now.Add(-30 * time.Minute)
	to := now

	result, err := s.agg.Aggregation(context.Background(), from, to)
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), result)
}

func (s *ESIntegrationSuite) TestAggregation_DocCountCorrect() {
	now := time.Now()
	for i := 0; i < 5; i++ {
		s.indexHeartbeat("s1", "on", now.Add(-time.Duration(i)*time.Minute))
	}
	s.client.Indices.Refresh()
	time.Sleep(500 * time.Millisecond)

	result, err := s.agg.Aggregation(context.Background(), now.Add(-time.Hour), now.Add(time.Minute))
	assert.NoError(s.T(), err)
	require.Len(s.T(), result, 1)
	assert.Equal(s.T(), int64(5), result[0].DocCount)
	assert.Equal(s.T(), fmt.Sprintf("s%d", 1), result[0].ServerID)
}
