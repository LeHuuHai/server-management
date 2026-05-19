package repo

import (
	"context"

	"github.com/LeHuuHai/server-management/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ServerRepo struct {
	db *gorm.DB
}

func (r *ServerRepo) Create(ctx context.Context, s *model.Server) error {
	return r.db.WithContext(ctx).
		Create(s).
		Error
}

func (r *ServerRepo) Update(ctx context.Context, id string, fields map[string]any) error {
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

func (r *ServerRepo) Delete(ctx context.Context, id string) error {
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

func (r *ServerRepo) List(ctx context.Context, filter model.ListServerFilter) (*model.ListServerResult, error) {
	var servers []model.Server
	var total int64

	query := r.db.WithContext(ctx).
		Model(&model.Server{}).
		Where("is_deleted = false")

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	err := query.
		Order(clause.OrderByColumn{
			Column: clause.Column{Name: string(filter.SortField)},
			Desc:   filter.Desc,
		}).
		Offset(filter.From).
		Limit(filter.To - filter.From).
		Find(&servers).Error

	return &model.ListServerResult{
		Servers: servers,
		Total:   int(total),
	}, err
}

func (r *ServerRepo) CreateBatch(ctx context.Context, servers []model.Server) (*model.CreateBatchServerResult, error) {
	res := &model.CreateBatchServerResult{
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

func NewServerRepository(db *gorm.DB) *ServerRepo {
	return &ServerRepo{db: db}
}
