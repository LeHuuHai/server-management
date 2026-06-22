package pg

import (
	"errors"
	"fmt"

	"github.com/LeHuuHai/server-management/microservices/auth-service/internal/model"
	apperr "github.com/LeHuuHai/server-management/microservices/pkg/apperr"
	"gorm.io/gorm"
)

type AccountRepo struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) *AccountRepo {
	return &AccountRepo{db: db}
}

func (r *AccountRepo) FindByUserName(userName string) (*model.Account, error) {
	var account model.Account
	err := r.db.Where("username = ?", userName).First(&account).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w", apperr.ErrRecordNotFound)
		}
		return nil, fmt.Errorf("%w: %v", apperr.ErrConnectPostgres, err)
	}
	return &account, nil
}

func (r *AccountRepo) FindByUserID(userID uint) (*model.Account, error) {
	var account model.Account
	err := r.db.Where("user_id = ?", userID).First(&account).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w", apperr.ErrRecordNotFound)
		}
		return nil, fmt.Errorf("%w: %v", apperr.ErrConnectPostgres, err)
	}
	return &account, nil
}
