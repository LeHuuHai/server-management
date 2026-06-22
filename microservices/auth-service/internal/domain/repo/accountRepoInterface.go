package repo

import "github.com/LeHuuHai/server-management/microservices/auth-service/internal/model"

type AccountRepoInterface interface {
	FindByUserName(userName string) (*model.Account, error)
	FindByUserID(userID uint) (*model.Account, error)
}
