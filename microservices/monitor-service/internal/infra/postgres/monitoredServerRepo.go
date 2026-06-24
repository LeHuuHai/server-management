package postgres

import (
	"context"
	"errors"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/repo"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/model"
	"gorm.io/gorm"
)

type MonitoredServerRepo struct {
	db *gorm.DB
}

func NewMonitoredServerRepository(db *gorm.DB) repo.MonitoredServerRepositoryInterface {
	return &MonitoredServerRepo{db: db}
}

func (r *MonitoredServerRepo) Create(ctx context.Context, s *model.MonitoredServer) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *MonitoredServerRepo) Update(ctx context.Context, s *model.MonitoredServer) error {
	// Versioning guard to prevent out-of-order writes
	res := r.db.WithContext(ctx).
		Model(&model.MonitoredServer{}).
		Where("server_id = ? AND version < ?", s.ServerID, s.Version).
		Updates(map[string]any{
			"server_name": s.ServerName,
			"ipv4":        s.IPv4,
			"version":     s.Version,
		})
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *MonitoredServerRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where("server_id = ?", id).
		Delete(&model.MonitoredServer{}).
		Error
}

func (r *MonitoredServerRepo) Get(ctx context.Context, id string) (*model.MonitoredServer, error) {
	var s model.MonitoredServer
	err := r.db.WithContext(ctx).Where("server_id = ?", id).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &s, err
}

func (r *MonitoredServerRepo) List(ctx context.Context) ([]model.MonitoredServer, error) {
	var list []model.MonitoredServer
	err := r.db.WithContext(ctx).Find(&list).Error
	return list, err
}
