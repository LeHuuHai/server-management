package postgres

import (
	"context"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/repo"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LiveStatusRepo struct {
	db *gorm.DB
}

func NewLiveStatusRepository(db *gorm.DB) repo.LiveStatusRepositoryInterface {
	return &LiveStatusRepo{db: db}
}

func (r *LiveStatusRepo) Create(ctx context.Context, s *model.LiveStatus) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *LiveStatusRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where("server_id = ?", id).
		Delete(&model.LiveStatus{}).
		Error
}

func (r *LiveStatusRepo) List(ctx context.Context) ([]model.LiveStatus, error) {
	var list []model.LiveStatus
	err := r.db.WithContext(ctx).Find(&list).Error
	return list, err
}

func (r *LiveStatusRepo) BulkUpdateLiveStatus(ctx context.Context, items []model.LiveStatus) error {
	if len(items) == 0 {
		return nil
	}

	// Single batch GORM bulk upsert (insert or update status, last_ping_at, last_heartbeat_at on conflict)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "server_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "last_ping_at", "last_heartbeat_at"}),
	}).Create(&items).Error
}
