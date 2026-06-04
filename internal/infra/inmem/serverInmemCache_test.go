package inmem

import (
	"context"
	"testing"
	"time"

	"github.com/LeHuuHai/server-management/internal/model"
)

// mockRepo minimal implementation for tests
type mockRepo struct {
	metadata []model.ServerMetadata
	err      error
}

func (m *mockRepo) AllMetadata(ctx context.Context) ([]model.ServerMetadata, error) {
	return m.metadata, m.err
}
func (m *mockRepo) Create(ctx context.Context, s *model.Server) error { return nil }
func (m *mockRepo) Update(ctx context.Context, id string, fields map[string]any) (*model.Server, error) {
	return nil, nil
}
func (m *mockRepo) Delete(ctx context.Context, id string) error { return nil }
func (m *mockRepo) List(ctx context.Context, filter model.ListServerFilter) (*model.ListServerResult, error) {
	return nil, nil
}
func (m *mockRepo) CreateBatch(ctx context.Context, servers []model.Server) (*model.CreateBatchServerResult, error) {
	return nil, nil
}
func (m *mockRepo) BulkUpdateServers(ctx context.Context, items []model.Server) error { return nil }

func newTestCache(t *testing.T, initial ...model.ServerMetadata) *serverInmemCache {
	t.Helper()
	repo := &mockRepo{metadata: initial}
	c, err := NewServerInmemCache(context.Background(), repo)
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}
	return c
}

func TestCreate_AddsEntry(t *testing.T) {
	c := newTestCache(t)
	c.Create(context.Background(), model.ServerMetadata{ServerID: "s1", ServerName: "web-01", IPv4: "10.0.0.1"})

	list := c.List(context.Background())
	if len(list) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(list))
	}
	if list[0].ServerID != "s1" {
		t.Errorf("expected s1, got %s", list[0].ServerID)
	}
}

func TestCreate_IgnoresExistingID(t *testing.T) {
	c := newTestCache(t, model.ServerMetadata{
		ServerID:   "s1",
		ServerName: "old",
		IPv4:       "10.0.0.1",
	})

	c.Create(context.Background(), model.ServerMetadata{
		ServerID:   "s1",
		ServerName: "new",
		IPv4:       "10.0.0.2",
	})

	list := c.List(context.Background())
	if len(list) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(list))
	}

	if list[0].ServerName != "old" {
		t.Errorf("expected existing server name to remain old, got %s", list[0].ServerName)
	}
	if list[0].IPv4 != "10.0.0.1" {
		t.Errorf("expected existing ipv4 to remain 10.0.0.1, got %s", list[0].IPv4)
	}
}

func TestUpdate_ChangesNameAndIP(t *testing.T) {
	c := newTestCache(t, model.ServerMetadata{ServerID: "s1", ServerName: "old", IPv4: "10.0.0.1"})

	c.Update(context.Background(), model.ServerMetadata{ServerID: "s1", ServerName: "new", IPv4: "10.0.0.2"})

	list := c.List(context.Background())
	if len(list) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(list))
	}
	if list[0].ServerName != "new" || list[0].IPv4 != "10.0.0.2" {
		t.Errorf("unexpected values: %+v", list[0])
	}
}

func TestUpdate_IgnoresUnknownServer(t *testing.T) {
	c := newTestCache(t)

	c.Update(context.Background(), model.ServerMetadata{
		ServerID:   "unknown",
		ServerName: "web-01",
		IPv4:       "10.0.0.1",
	})

	list := c.List(context.Background())
	if len(list) != 0 {
		t.Fatalf("expected no entries, got %d", len(list))
	}
}

func TestUpdate_PreservesHeartbeat(t *testing.T) {
	now := time.Now()
	initial := model.ServerMetadata{
		ServerID:        "s1",
		ServerName:      "old",
		IPv4:            "10.0.0.1",
		LastHeartbeatAt: &now,
	}
	c := newTestCache(t, initial)

	c.Update(context.Background(), model.ServerMetadata{ServerID: "s1", ServerName: "new", IPv4: "10.0.0.2"})

	list := c.List(context.Background())
	if list[0].LastHeartbeatAt == nil {
		t.Error("expected heartbeat timestamp to be preserved after update")
	}
}

func TestDelete_RemovesEntry(t *testing.T) {
	c := newTestCache(t,
		model.ServerMetadata{ServerID: "s1"},
		model.ServerMetadata{ServerID: "s2"},
	)

	c.Delete(context.Background(), "s1")

	list := c.List(context.Background())
	if len(list) != 1 {
		t.Fatalf("expected 1 entry after delete, got %d", len(list))
	}
	if list[0].ServerID != "s2" {
		t.Errorf("expected s2 to remain, got %s", list[0].ServerID)
	}
}

func TestDelete_IgnoresUnknownServer(t *testing.T) {
	c := newTestCache(t,
		model.ServerMetadata{ServerID: "s1", ServerName: "web-01", IPv4: "10.0.0.1"},
	)

	c.Delete(context.Background(), "unknown")

	list := c.List(context.Background())
	if len(list) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(list))
	}

	if list[0].ServerID != "s1" {
		t.Errorf("expected s1 to remain, got %s", list[0].ServerID)
	}
}

func TestBatchCreate_AddsMultiple(t *testing.T) {
	c := newTestCache(t)
	items := []model.ServerMetadata{
		{ServerID: "s1", ServerName: "web-01"},
		{ServerID: "s2", ServerName: "web-02"},
	}
	c.BatchCreate(context.Background(), items)

	list := c.List(context.Background())
	if len(list) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(list))
	}
}

func TestBatchCreate_IgnoresExistingID(t *testing.T) {
	c := newTestCache(t, model.ServerMetadata{
		ServerID:   "s1",
		ServerName: "old",
		IPv4:       "10.0.0.1",
	})

	c.BatchCreate(context.Background(), []model.ServerMetadata{
		{ServerID: "s1", ServerName: "new", IPv4: "10.0.0.2"},
		{ServerID: "s2", ServerName: "web-02", IPv4: "10.0.0.3"},
	})

	list := c.List(context.Background())
	if len(list) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(list))
	}

	byID := make(map[string]model.ServerMetadata)
	for _, item := range list {
		byID[item.ServerID] = item
	}

	if byID["s1"].ServerName != "old" {
		t.Errorf("expected s1 name to remain old, got %s", byID["s1"].ServerName)
	}
	if byID["s1"].IPv4 != "10.0.0.1" {
		t.Errorf("expected s1 ipv4 to remain 10.0.0.1, got %s", byID["s1"].IPv4)
	}
	if _, ok := byID["s2"]; !ok {
		t.Error("expected s2 to be created")
	}
}

func TestBatchUpdateHeartbeat_UpdatesExisting(t *testing.T) {
	c := newTestCache(t, model.ServerMetadata{
		ServerID:   "s1",
		ServerName: "web-01",
		IPv4:       "10.0.0.1",
	})

	now := time.Now()
	c.BatchUpdateHeartbeat(context.Background(), []model.ServerMetadata{
		{ServerID: "s1", LastHeartbeatAt: &now},
	})

	list := c.List(context.Background())

	if len(list) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(list))
	}

	got := list[0]
	if got.LastHeartbeatAt == nil {
		t.Fatal("expected heartbeat to be set")
	}
	if !got.LastHeartbeatAt.Equal(now) {
		t.Errorf("expected heartbeat %v, got %v", now, got.LastHeartbeatAt)
	}
	if got.ServerName != "web-01" {
		t.Errorf("expected server name to remain web-01, got %s", got.ServerName)
	}
	if got.IPv4 != "10.0.0.1" {
		t.Errorf("expected ipv4 to remain 10.0.0.1, got %s", got.IPv4)
	}
}

func TestBatchUpdateHeartbeat_IgnoresUnknownServer(t *testing.T) {
	c := newTestCache(t)

	now := time.Now()
	// Should not panic for unknown server ID
	c.BatchUpdateHeartbeat(context.Background(), []model.ServerMetadata{
		{ServerID: "unknown", LastHeartbeatAt: &now},
	})

	list := c.List(context.Background())

	if len(list) != 0 {
		t.Fatalf("expected 0 entry, got %d", len(list))
	}
}

func TestSync_ReplacesCache(t *testing.T) {
	repo := &mockRepo{
		metadata: []model.ServerMetadata{
			{ServerID: "s1"}, {ServerID: "s2"},
		},
	}

	c, err := NewServerInmemCache(context.Background(), repo)
	if err != nil {
		t.Fatal(err)
	}

	c.Create(context.Background(), model.ServerMetadata{ServerID: "s3"})

	repo.metadata = []model.ServerMetadata{{ServerID: "s99"}}
	err = c.Sync(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	list := c.List(context.Background())
	if len(list) != 1 {
		t.Fatalf("expected 1 entry after sync, got %d", len(list))
	}

	if list[0].ServerID != "s99" {
		t.Errorf("expected only s99 after sync, got %+v", list)
	}
}

func TestNewServerInmemCache_RepoError(t *testing.T) {
	repo := &mockRepo{err: &repoError{}}
	_, err := NewServerInmemCache(context.Background(), repo)
	if err == nil {
		t.Error("expected error when repo.AllMetadata fails")
	}
}

type repoError struct{}

func (e *repoError) Error() string { return "repo error" }
