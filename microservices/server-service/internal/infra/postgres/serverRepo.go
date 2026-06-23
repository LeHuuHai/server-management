package pg

import (
	"context"
	"errors"
	"fmt"
	"strings"

	apperr "github.com/LeHuuHai/server-management/microservices/pkg/apperr"
	"github.com/LeHuuHai/server-management/microservices/server-service/internal/model"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ServerRepo struct {
	db *gorm.DB
}

func (r *ServerRepo) Create(ctx context.Context, s *model.Server) error {
	s.Status = model.StatusUnknown
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

func (r *ServerRepo) Update(ctx context.Context, id string, fields map[string]any) (*model.Server, error) {
	var updated model.Server

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

func (r *ServerRepo) List(ctx context.Context, filter model.ListServerFilter) (*model.ListServerResult, error) {
	var servers []model.Server
	var total int64

	db := getDB(ctx, r.db)
	query := db.WithContext(ctx).
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
		s.Status = model.StatusUnknown
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

func (r *ServerRepo) AllMetadata(ctx context.Context) ([]model.ServerMetadata, error) {
	var result []model.ServerMetadata

	db := getDB(ctx, r.db)
	err := db.WithContext(ctx).
		Model(&model.Server{}).
		Select("server_id", "server_name", "ipv4").
		Where("is_deleted = false").
		Find(&result).
		Error

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *ServerRepo) BulkUpdateServers(ctx context.Context, items []model.Server) error {
	var b strings.Builder
	b.WriteString(`
		UPDATE servers AS s
		SET
			status = v.status,
			last_ping_at = v.last_ping_at
		FROM (VALUES `)

	args := make([]any, 0, len(items)*3)

	for i, it := range items {
		if i > 0 {
			b.WriteString(",")
		}

		// QUAN TRá»ŒNG: Ã‰p kiá»ƒu á»Ÿ dÃ²ng Ä‘áº§u tiÃªn Ä‘á»ƒ Postgres nháº­n diá»‡n Ä‘Ãºng kiá»ƒu dá»¯ liá»‡u
		if i == 0 {
			// Ã‰p kiá»ƒu id thÃ nh varchar/text (hoáº·c int/uuid tÃ¹y DB cá»§a báº¡n), status thÃ nh varchar, vÃ  last_ping_at thÃ nh timestamp
			b.WriteString(fmt.Sprintf("($%d::varchar, $%d::varchar, $%d::timestamp)", i*3+1, i*3+2, i*3+3))
		} else {
			b.WriteString(fmt.Sprintf("($%d,$%d,$%d)", i*3+1, i*3+2, i*3+3))
		}

		args = append(args,
			it.ServerID,
			it.Status,
			it.LastPingAt,
		)
	}

	b.WriteString(`
		) AS v(id, status, last_ping_at)
		WHERE s.server_id = v.id
		`)

	db := getDB(ctx, r.db)
	res := db.WithContext(ctx).Exec(b.String(), args...)
	return res.Error
}

func NewServerRepository(db *gorm.DB) *ServerRepo {
	return &ServerRepo{db: db}
}
