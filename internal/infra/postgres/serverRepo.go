package pg

import (
	"context"
	"errors"

	apperr "github.com/LeHuuHai/server-management/internal/error"
	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type serverRepo struct {
	db *gorm.DB
}

func (r *serverRepo) Create(ctx context.Context, s *model.Server) error {
	err := r.db.WithContext(ctx).
		Create(s).
		Error
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			// unique_violation
			case "23505":
				return apperr.ErrDuplicateServer
			}
		}
		return err
	}
	return nil
}

func (r *serverRepo) Update(ctx context.Context, id string, fields map[string]any) (*model.Server, error) {
	var updated model.Server

	res := r.db.WithContext(ctx).
		Model(&updated).
		Clauses(clause.Returning{}).
		Where("server_id = ? AND is_deleted = false", id).
		Updates(fields)

	if res.Error != nil {
		return nil, res.Error
	}

	if res.RowsAffected == 0 {
		return nil, apperr.ErrRecordNotFound
	}

	return &updated, nil
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
		return apperr.ErrRecordNotFound
	}

	return nil
}

func (r *serverRepo) List(ctx context.Context, filter model.ListServerFilter) (*model.ListServerResult, error) {
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

func (r *serverRepo) CreateBatch(ctx context.Context, servers []model.Server) (*model.CreateBatchServerResult, error) {
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

func (r *serverRepo) AllMetadata(ctx context.Context) ([]model.ServerMetadata, error) {
	var result []model.ServerMetadata

	err := r.db.WithContext(ctx).
		Model(&model.Server{}).
		Select("server_id", "server_name", "ipv4").
		Find(&result).
		Error

	if err != nil {
		return nil, err
	}

	return result, nil
}

func NewServerRepository(db *gorm.DB) *serverRepo {
	return &serverRepo{db: db}
}
