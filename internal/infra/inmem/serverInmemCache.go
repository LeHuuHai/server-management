package inmem

import (
	"context"
	"sync"

	"github.com/LeHuuHai/server-management/internal/model"
)

type serverInmemCache struct {
	servers map[string]*model.ServerMetadata
	mu      sync.Mutex
}

func NewServerInmemCache() *serverInmemCache {
	return &serverInmemCache{
		servers: make(map[string]*model.ServerMetadata),
		mu:      sync.Mutex{},
	}
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
