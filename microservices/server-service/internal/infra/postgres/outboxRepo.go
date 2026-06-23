package pg

import (
	"context"

	"github.com/LeHuuHai/server-management/microservices/server-service/internal/model"
	"gorm.io/gorm"
)

type OutboxRepo struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) *OutboxRepo {
	return &OutboxRepo{db: db}
}

func (r *OutboxRepo) CreateEvent(ctx context.Context, event *model.OutboxEvent) error {
	db := getDB(ctx, r.db)
	return db.WithContext(ctx).Create(event).Error
}

func (r *OutboxRepo) FetchPendingEvents(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
	var events []model.OutboxEvent
	db := getDB(ctx, r.db)

	err := db.WithContext(ctx).
		Where("status = ?", model.OutboxStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Find(&events).Error

	return events, err
}

func (r *OutboxRepo) DeleteEvents(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	db := getDB(ctx, r.db)
	return db.WithContext(ctx).Where("id IN ?", ids).Delete(&model.OutboxEvent{}).Error
}
