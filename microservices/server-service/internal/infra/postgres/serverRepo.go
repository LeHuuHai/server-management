package pg

import (
	"context"
	"errors"

	apperr "github.com/LeHuuHai/server-management/microservices/pkg/apperr"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/model"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ServerRepo struct {
	db *gorm.DB
}

func (r *ServerRepo) Create(ctx context.Context, s *model.ServerProfile) error {
	db := getDB(ctx, r.db)
	err := db.WithContext(ctx).
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

func (r *ServerRepo) Update(ctx context.Context, id string, fields map[string]any) (*model.ServerProfile, error) {
	var updated model.ServerProfile

	fields["version"] = gorm.Expr("version + 1")

	db := getDB(ctx, r.db)
	res := db.WithContext(ctx).
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

func (r *ServerRepo) Delete(ctx context.Context, id string) error {
	db := getDB(ctx, r.db)
	res := db.WithContext(ctx).
		Model(&model.ServerProfile{}).
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

func (r *ServerRepo) List(ctx context.Context, filter model.ListServerFilter) (*model.ListServerResult, error) {
	var servers []model.ServerProfile
	var total int64

	db := getDB(ctx, r.db)
	query := db.WithContext(ctx).
		Model(&model.ServerProfile{}).
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

func (r *ServerRepo) CreateBatch(ctx context.Context, servers []model.ServerProfile) (*model.CreateBatchServerResult, error) {
	res := &model.CreateBatchServerResult{
		Success:    make([]string, 0),
		Failed:     make([]string, 0),
		SuccessCnt: 0,
		FailedCnt:  0,
	}

	for _, s := range servers {
		db := getDB(ctx, r.db)
		err := db.WithContext(ctx).
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
