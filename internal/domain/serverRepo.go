package domain

import (
	"context"
	"time"

	"github.com/LeHuuHai/server-management/internal/model"
	"gorm.io/gorm"
)

type serverRepo struct {
	db *gorm.DB
}

func (r *serverRepo) Create(ctx context.Context, s *model.Server) error {
	return r.db.WithContext(ctx).
		Create(s).
		Error
}

func (r *serverRepo) Update(ctx context.Context, id string, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}

	allowed := map[string]bool{
		"server_name": true,
		"ipv4":        true,
	}

	safeFields := make(map[string]any)
	for k, v := range fields {
		if allowed[k] {
			safeFields[k] = v
		}
	}
	safeFields["last_updated"] = time.Now()
	if len(safeFields) == 1 {
		return nil
	}

	res := r.db.WithContext(ctx).
		Model(&model.Server{}).
		Where("server_id = ? AND is_deleted = false", id).
		Updates(safeFields)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *serverRepo) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).
		Model(&model.Server{}).
		Where("server_id = ? AND is_deleted = false", id).
		Update("is_deleted", true)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *serverRepo) List(ctx context.Context, filter model.ListServerFilter) ([]model.Server, int, error) {
	var (
		servers []model.Server
		total   int64
	)

	query := r.db.WithContext(ctx).
		Model(&model.Server{}).
		Where("is_deleted = false")

	// count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// sorting
	sortField := "created_time"
	if filter.SortField != "" {
		sortField = filter.SortField
	}

	sortOrder := "desc"
	if filter.SortOrder == "ASC" {
		sortOrder = "asc"
	}

	limit := filter.To - filter.From
	if limit <= 0 {
		limit = 10
	}

	err := query.
		Order(sortField + " " + sortOrder).
		Offset(filter.From).
		Limit(limit).
		Find(&servers).Error

	return servers, int(total), err
}

func (r *serverRepo) CreateBatch(ctx context.Context, servers []model.Server) (*BatchResult, error) {
	res := &BatchResult{
		Success:     make([]string, 0),
		Failed:      make([]string, 0),
		Success_cnt: 0,
		Failed_cnt:  0,
	}

	for _, s := range servers {
		err := r.db.WithContext(ctx).
			Create(&s).Error

		if err != nil {
			res.Failed = append(res.Failed, s.ServerID)
			continue
		}

		res.Success = append(res.Success, s.ServerID)
	}

	res.Success_cnt = len(res.Success)
	res.Failed_cnt = len(res.Failed)

	return res, nil
}

func NewServerRepository(db *gorm.DB) ServerRepository {
	return &serverRepo{db: db}
}
