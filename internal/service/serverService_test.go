package service_test

import (
	"context"
	"errors"
	"testing"

	apperr "github.com/LeHuuHai/server-management/internal/error"
	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/LeHuuHai/server-management/internal/service"
)

// ---- helpers ----

type mockServerRepo struct {
	servers map[string]model.Server

	createErr      error
	updateErr      error
	deleteErr      error
	listErr        error
	createBatchErr error
	allMetadataErr error
	bulkUpdateErr  error

	listResult        *model.ListServerResult
	createBatchResult *model.CreateBatchServerResult
	metadata          []model.ServerMetadata

	lastListFilter  model.ListServerFilter
	lastUpdateID    string
	lastUpdateFields map[string]any
	batchCreated    []model.Server
}

func (m *mockServerRepo) Create(ctx context.Context, s *model.Server) error {
	if m.createErr != nil {
		return m.createErr
	}
	if m.servers == nil {
		m.servers = make(map[string]model.Server)
	}
	m.servers[s.ServerID] = *s
	return nil
}

func (m *mockServerRepo) Update(ctx context.Context, id string, fields map[string]any) (*model.Server, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	if m.servers == nil {
		m.servers = make(map[string]model.Server)
	}

	server := m.servers[id]
	if server.ServerID == "" {
		server.ServerID = id
	}
	if name, ok := fields["server_name"].(string); ok {
		server.ServerName = name
	}
	if ipv4, ok := fields["ipv4"].(string); ok {
		server.IPv4 = ipv4
	}
	m.servers[id] = server

	m.lastUpdateID = id
	m.lastUpdateFields = fields
	return &server, nil
}

func (m *mockServerRepo) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.servers, id)
	return nil
}

func (m *mockServerRepo) List(ctx context.Context, filter model.ListServerFilter) (*model.ListServerResult, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	m.lastListFilter = filter
	if m.listResult != nil {
		return m.listResult, nil
	}

	servers := make([]model.Server, 0, len(m.servers))
	for _, server := range m.servers {
		servers = append(servers, server)
	}
	return &model.ListServerResult{Servers: servers, Total: len(servers)}, nil
}

func (m *mockServerRepo) CreateBatch(ctx context.Context, servers []model.Server) (*model.CreateBatchServerResult, error) {
	if m.createBatchErr != nil {
		return nil, m.createBatchErr
	}
	if m.servers == nil {
		m.servers = make(map[string]model.Server)
	}
	m.batchCreated = servers
	for _, server := range servers {
		m.servers[server.ServerID] = server
	}
	if m.createBatchResult != nil {
		return m.createBatchResult, nil
	}
	return &model.CreateBatchServerResult{
		SuccessCnt: len(servers),
		Failed:     []string{},
		FailedCnt:  0,
	}, nil
}

func (m *mockServerRepo) AllMetadata(ctx context.Context) ([]model.ServerMetadata, error) {
	if m.allMetadataErr != nil {
		return nil, m.allMetadataErr
	}
	return m.metadata, nil
}

func (m *mockServerRepo) BulkUpdateServers(ctx context.Context, items []model.Server) error {
	if m.bulkUpdateErr != nil {
		return m.bulkUpdateErr
	}
	if m.servers == nil {
		m.servers = make(map[string]model.Server)
	}
	for _, item := range items {
		m.servers[item.ServerID] = item
	}
	return nil
}

type mockServerCache struct {
	servers map[string]model.ServerMetadata
	syncErr error
}

func (m *mockServerCache) Create(ctx context.Context, s model.ServerMetadata) {
	if m.servers == nil {
		m.servers = make(map[string]model.ServerMetadata)
	}
	m.servers[s.ServerID] = model.ServerMetadata{
		ServerID:   s.ServerID,
		ServerName: s.ServerName,
		IPv4:       s.IPv4,
	}
}

func (m *mockServerCache) Update(ctx context.Context, s model.ServerMetadata) {
	if m.servers == nil {
		m.servers = make(map[string]model.ServerMetadata)
	}
	current := m.servers[s.ServerID]
	m.servers[s.ServerID] = model.ServerMetadata{
		ServerID:         s.ServerID,
		ServerName:       s.ServerName,
		IPv4:             s.IPv4,
		LastHeartbeatAt:  current.LastHeartbeatAt,
	}
}

func (m *mockServerCache) BatchUpdateHeartbeat(ctx context.Context, items []model.ServerMetadata) {
	for _, item := range items {
		server, ok := m.servers[item.ServerID]
		if !ok {
			continue
		}
		server.LastHeartbeatAt = item.LastHeartbeatAt
		m.servers[item.ServerID] = server
	}
}

func (m *mockServerCache) Delete(ctx context.Context, id string) {
	delete(m.servers, id)
}

func (m *mockServerCache) BatchCreate(ctx context.Context, items []model.ServerMetadata) {
	if m.servers == nil {
		m.servers = make(map[string]model.ServerMetadata)
	}
	for _, item := range items {
		m.servers[item.ServerID] = model.ServerMetadata{
			ServerID:   item.ServerID,
			ServerName: item.ServerName,
			IPv4:       item.IPv4,
		}
	}
}

func (m *mockServerCache) List(ctx context.Context) []model.ServerMetadata {
	servers := make([]model.ServerMetadata, 0, len(m.servers))
	for _, server := range m.servers {
		servers = append(servers, server)
	}
	return servers
}

func (m *mockServerCache) Sync(ctx context.Context) error {
	return m.syncErr
}

func newServerService(repo *mockServerRepo, cache *mockServerCache) *service.ServerService {
	return service.NewServerService(repo, cache)
}

// ---- CreateServer ----

func TestCreateServer_Success(t *testing.T) {
	repo := &mockServerRepo{}
	cache := &mockServerCache{}
	svc := newServerService(repo, cache)

	server := &model.Server{ServerID: "s1", ServerName: "web-01", IPv4: "192.168.1.1"}
	result, err := svc.CreateServer(context.Background(), server)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ServerID != "s1" {
		t.Errorf("expected server id s1, got %s", result.ServerID)
	}
	if len(cache.servers) != 1 {
		t.Fatalf("expected cache to have one server, got %d", len(cache.servers))
	}
	cached := cache.servers["s1"]
	if cached.ServerID != "s1" || cached.ServerName != "web-01" || cached.IPv4 != "192.168.1.1" {
		t.Errorf("expected cache to have correct server, got %+v", cached)
	}
}

func TestCreateServer_InvalidIPv4(t *testing.T) {
	cases := []string{"not-an-ip", "999.999.999.999", "::1", ""}
	repo := &mockServerRepo{}
	cache := &mockServerCache{}
	svc := newServerService(repo, cache)

	for _, ip := range cases {
		server := &model.Server{ServerID: "s1", ServerName: "web-01", IPv4: ip}
		_, err := svc.CreateServer(context.Background(), server)
		if !errors.Is(err, apperr.ErrInvalidIP) {
			t.Errorf("ip=%q: expected ErrInvalidIP, got %v", ip, err)
		}
	}
	if len(cache.servers) != 0 {
		t.Error("cache should not be updated for invalid ip")
	}
}

func TestCreateServer_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	repo := &mockServerRepo{createErr: repoErr}
	cache := &mockServerCache{}
	svc := newServerService(repo, cache)

	_, err := svc.CreateServer(context.Background(), &model.Server{
		ServerID:   "s1",
		ServerName: "web-01",
		IPv4:       "10.0.0.1",
	})
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repoErr, got %v", err)
	}
	if len(cache.servers) != 0 {
		t.Error("cache should not be updated on repo error")
	}
}

func TestCreateServer_DuplicateID(t *testing.T) {
	repo := &mockServerRepo{createErr: apperr.ErrDuplicateServer}
	cache := &mockServerCache{}
	svc := newServerService(repo, cache)

	_, err := svc.CreateServer(context.Background(), &model.Server{
		ServerID:   "s1",
		ServerName: "web-01",
		IPv4:       "10.0.0.1",
	})
	if !errors.Is(err, apperr.ErrDuplicateServer) {
		t.Errorf("expected ErrDuplicateServer, got %v", err)
	}
	if len(cache.servers) != 0 {
		t.Error("cache should not be updated on duplicate id")
	}
}

// ---- ListServer ----

func TestListServer_Success(t *testing.T) {
	want := &model.ListServerResult{}
	repo := &mockServerRepo{listResult: want}
	svc := newServerService(repo, &mockServerCache{})

	filter := model.ListServerFilter{From: 0, To: 10, SortField: model.SortByName}
	result, err := svc.ListServer(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != want {
		t.Error("unexpected result")
	}
}

func TestListServer_InvalidSortField(t *testing.T) {
	svc := newServerService(&mockServerRepo{}, &mockServerCache{})
	filter := model.ListServerFilter{From: 0, To: 10, SortField: "unknown_field"}
	_, err := svc.ListServer(context.Background(), filter)
	if !errors.Is(err, apperr.ErrInvalidSort) {
		t.Errorf("expected ErrInvalidSort, got %v", err)
	}
}

func TestListServer_InvalidPagination(t *testing.T) {
	svc := newServerService(&mockServerRepo{}, &mockServerCache{})

	cases := []model.ListServerFilter{
		{From: -1, To: 10, SortField: model.SortByName},
		{From: 10, To: 5, SortField: model.SortByName},
		{From: 0, To: 0, SortField: model.SortByName},
	}

	for _, f := range cases {
		_, err := svc.ListServer(context.Background(), f)
		if !errors.Is(err, apperr.ErrInvalidPagination) {
			t.Errorf("filter=%+v: expected ErrInvalidPagination, got %v", f, err)
		}
	}
}

func TestListServer_CapsToMaxPageSize(t *testing.T) {
	repo := &mockServerRepo{}
	svc := newServerService(repo, &mockServerCache{})

	// Request 200 items; should be capped at 100
	_, err := svc.ListServer(context.Background(), model.ListServerFilter{
		From: 0, To: 200, SortField: model.SortByCreatedAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	if repo.lastListFilter.To-repo.lastListFilter.From > 100 {
		t.Errorf("expected page size <= 100, got %d", repo.lastListFilter.To-repo.lastListFilter.From)
	}
}

func TestListServer_RepoError(t *testing.T) {
	repoErr := errors.New("list failed")
	repo := &mockServerRepo{listErr: repoErr}
	svc := newServerService(repo, &mockServerCache{})

	_, err := svc.ListServer(context.Background(), model.ListServerFilter{
		From:      0,
		To:        10,
		SortField: model.SortByName,
	})
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repoErr, got %v", err)
	}
}

// ---- UpdateServer ----

func TestUpdateServer_Success(t *testing.T) {
	repo := &mockServerRepo{}
	cache := &mockServerCache{}
	svc := newServerService(repo, cache)

	result, err := svc.UpdateServer(context.Background(), &model.Server{
		ServerID: "s1", ServerName: "new-name", IPv4: "10.0.0.2",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ServerName != "new-name" {
		t.Errorf("expected new-name, got %s", result.ServerName)
	}
	if len(cache.servers) != 1 {
		t.Fatalf("expected cache to have one server, got %d", len(cache.servers))
	}
	cached := cache.servers["s1"]
	if cached.ServerName != "new-name" || cached.IPv4 != "10.0.0.2" {
		t.Errorf("expected cache to have updated server, got %+v", cached)
	}
}

func TestUpdateServer_InvalidIPv4(t *testing.T) {
	svc := newServerService(&mockServerRepo{}, &mockServerCache{})
	_, err := svc.UpdateServer(context.Background(), &model.Server{
		ServerID: "s1", IPv4: "bad-ip",
	})
	if !errors.Is(err, apperr.ErrInvalidIP) {
		t.Errorf("expected ErrInvalidIP, got %v", err)
	}
}

func TestUpdateServer_RepoError(t *testing.T) {
	repoErr := errors.New("update failed")
	repo := &mockServerRepo{updateErr: repoErr}
	cache := &mockServerCache{}
	svc := newServerService(repo, cache)

	_, err := svc.UpdateServer(context.Background(), &model.Server{ServerID: "s1"})
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repoErr, got %v", err)
	}
	if len(cache.servers) != 0 {
		t.Error("cache should not be updated on repo error")
	}
}

func TestUpdateServer_UnknownID(t *testing.T) {
	repo := &mockServerRepo{updateErr: apperr.ErrRecordNotFound}
	cache := &mockServerCache{}
	svc := newServerService(repo, cache)

	_, err := svc.UpdateServer(context.Background(), &model.Server{
		ServerID:   "unknown",
		ServerName: "new-name",
		IPv4:       "10.0.0.2",
	})
	if !errors.Is(err, apperr.ErrRecordNotFound) {
		t.Errorf("expected ErrRecordNotFound, got %v", err)
	}
	if len(cache.servers) != 0 {
		t.Error("cache should not be updated on unknown id")
	}
}

// ---- DeleteServer ----

func TestDeleteServer_Success(t *testing.T) {
	repo := &mockServerRepo{}
	cache := &mockServerCache{
		servers: map[string]model.ServerMetadata{
			"s1": {ServerID: "s1", ServerName: "web-01", IPv4: "10.0.0.1"},
		},
	}
	svc := newServerService(repo, cache)

	err := svc.DeleteServer(context.Background(), "s1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cache.servers["s1"]; ok {
		t.Error("expected server to be deleted from cache")
	}
}

func TestDeleteServer_RepoError(t *testing.T) {
	repoErr := errors.New("delete failed")
	repo := &mockServerRepo{deleteErr: repoErr}
	cache := &mockServerCache{}
	svc := newServerService(repo, cache)

	err := svc.DeleteServer(context.Background(), "s1")
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repoErr, got %v", err)
	}
	if len(cache.servers) != 0 {
		t.Error("cache should not be updated on repo error")
	}
}

func TestDeleteServer_UnknownID(t *testing.T) {
	repo := &mockServerRepo{deleteErr: apperr.ErrRecordNotFound}
	cache := &mockServerCache{
		servers: map[string]model.ServerMetadata{
			"s1": {ServerID: "s1", ServerName: "web-01", IPv4: "10.0.0.1"},
		},
	}
	svc := newServerService(repo, cache)

	err := svc.DeleteServer(context.Background(), "unknown")
	if !errors.Is(err, apperr.ErrRecordNotFound) {
		t.Errorf("expected ErrRecordNotFound, got %v", err)
	}
	if _, ok := cache.servers["s1"]; !ok {
		t.Error("cache should not be updated on unknown id")
	}
}

// ---- ImportServer ----

func TestImportServer_AllValid(t *testing.T) {
	batchResult := &model.CreateBatchServerResult{SuccessCnt: 2, Failed: []string{}, FailedCnt: 0}
	repo := &mockServerRepo{createBatchResult: batchResult}
	cache := &mockServerCache{}
	svc := newServerService(repo, cache)

	imports := []model.ServerImport{
		{ServerID: "s1", ServerName: "web-01", IPv4: "10.0.0.1"},
		{ServerID: "s2", ServerName: "web-02", IPv4: "192.168.0.1"},
	}
	result, err := svc.ImportServer(context.Background(), imports)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.FailedCnt != 0 {
		t.Errorf("expected 0 failures, got %d", result.FailedCnt)
	}
	if len(repo.batchCreated) != 2 {
		t.Errorf("expected 2 valid servers sent to repo, got %d", len(repo.batchCreated))
	}
	if len(cache.servers) != 2 {
		t.Errorf("expected 2 servers cached, got %d", len(cache.servers))
	}
}

func TestImportServer_SomeInvalidIPs(t *testing.T) {
	batchResult := &model.CreateBatchServerResult{SuccessCnt: 1, Failed: []string{}, FailedCnt: 0}
	repo := &mockServerRepo{createBatchResult: batchResult}
	cache := &mockServerCache{}
	svc := newServerService(repo, cache)

	imports := []model.ServerImport{
		{ServerID: "s1", ServerName: "web-01", IPv4: "10.0.0.1"},
		{ServerID: "bad", ServerName: "bad-server", IPv4: "not-an-ip"},
		{ServerID: "s3", ServerName: "web-03", IPv4: "::1"}, // IPv6, invalid
	}
	result, err := svc.ImportServer(context.Background(), imports)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.FailedCnt != 2 {
		t.Errorf("expected 2 failures, got %d", result.FailedCnt)
	}
	if len(repo.batchCreated) != 1 {
		t.Errorf("expected 1 valid server sent to repo, got %d", len(repo.batchCreated))
	}
}

func TestImportServer_RepoError(t *testing.T) {
	repoErr := errors.New("bulk insert failed")
	repo := &mockServerRepo{createBatchErr: repoErr}
	cache := &mockServerCache{}
	svc := newServerService(repo, cache)

	imports := []model.ServerImport{
		{ServerID: "s1", IPv4: "10.0.0.1"},
	}
	_, err := svc.ImportServer(context.Background(), imports)
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repoErr, got %v", err)
	}
	if len(cache.servers) != 0 {
		t.Error("cache should not be updated on repo error")
	}
}

func TestImportServer_DuplicateIDOrName(t *testing.T) {
	repo := &mockServerRepo{createBatchErr: apperr.ErrDuplicateServer}
	cache := &mockServerCache{}
	svc := newServerService(repo, cache)

	imports := []model.ServerImport{
		{ServerID: "s1", ServerName: "web-01", IPv4: "10.0.0.1"},
		{ServerID: "s1", ServerName: "web-02", IPv4: "10.0.0.2"},
	}
	_, err := svc.ImportServer(context.Background(), imports)
	if !errors.Is(err, apperr.ErrDuplicateServer) {
		t.Errorf("expected ErrDuplicateServer, got %v", err)
	}
	if len(cache.servers) != 0 {
		t.Error("cache should not be updated on duplicate id or name")
	}
}
