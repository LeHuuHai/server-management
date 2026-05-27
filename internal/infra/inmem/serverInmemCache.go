package inmem

import (
	"context"
	"sync"

	"github.com/LeHuuHai/server-management/internal/domain/repo"
	"github.com/LeHuuHai/server-management/internal/model"
)

type serverInmemCache struct {
	servers map[string]*model.ServerMetadata
	repo    repo.ServerRepositoryInterface
	mu      sync.Mutex
}

func NewServerInmemCache(ctx context.Context, r repo.ServerRepositoryInterface) (*serverInmemCache, error) {
	cache := serverInmemCache{
		servers: make(map[string]*model.ServerMetadata),
		mu:      sync.Mutex{},
		repo:    r,
	}
	err := cache.Sync(ctx)
	if err != nil {
		return nil, err
	}
	return &cache, nil
}

func (c *serverInmemCache) Create(ctx context.Context, s model.ServerMetadata) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.servers[s.ServerID] = &model.ServerMetadata{
		ServerID:   s.ServerID,
		ServerName: s.ServerName,
		IPv4:       s.IPv4,
	}
}

func (c *serverInmemCache) Update(ctx context.Context, s model.ServerMetadata) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.servers[s.ServerID] = &model.ServerMetadata{
		ServerID:   s.ServerID,
		ServerName: s.ServerName,
		IPv4:       s.IPv4,
	}
}

func (c *serverInmemCache) Delete(ctx context.Context, id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.servers, id)
}

func (c *serverInmemCache) BatchCreate(ctx context.Context, s []model.ServerMetadata) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range s {
		item := s[i]
		c.servers[item.ServerID] = &model.ServerMetadata{
			ServerID:   item.ServerID,
			ServerName: item.ServerName,
			IPv4:       item.IPv4,
		}
	}
}

func (c *serverInmemCache) List(ctx context.Context) []model.ServerMetadata {
	c.mu.Lock()
	defer c.mu.Unlock()
	servers := make([]model.ServerMetadata, 0, len(c.servers))
	for _, v := range c.servers {
		servers = append(servers, *v)
	}
	return servers
}

func (c *serverInmemCache) Sync(ctx context.Context) error {
	servers, err := c.repo.AllMetadata(ctx)
	if err != nil {
		return err
	}
	newServers := make(map[string]*model.ServerMetadata)
	for idx := range servers {
		s := &servers[idx]
		newServers[s.ServerID] = s
	}
	c.mu.Lock()
	c.servers = newServers
	c.mu.Unlock()
	return nil
}
