package domain

import (
	"context"

	"github.com/LeHuuHai/server-management/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	res := r.db.WithContext(ctx).
		Model(&model.Server{}).
		Where("server_id = ? AND is_deleted = false", id).
		Updates(fields)

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

	err := query.
		Order(clause.OrderByColumn{
			Column: clause.Column{Name: string(filter.SortField)},
			Desc:   filter.Desc,
		}).
		Offset(filter.From).
		Limit(filter.To - filter.From).
		Find(&servers).Error

	return servers, int(total), err
}

func (r *serverRepo) CreateBatch(ctx context.Context, servers []model.Server) (*BatchResult, error) {
	res := &BatchResult{
		Success:    make([]string, 0),
		Failed:     make([]string, 0),
		SuccessCnt: 0,
		FailedCnt:  0,
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

	res.SuccessCnt = len(res.Success)
	res.FailedCnt = len(res.Failed)

	return res, nil
}

func NewServerRepository(db *gorm.DB) ServerRepository {
	return &serverRepo{db: db}
}
