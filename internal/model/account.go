package model

import authdomain "github.com/LeHuuHai/server-management/internal/domain/auth"

type Account struct {
	ID       uint
	UserID   uint
	UserName string
	Password string
	Role     authdomain.Role
}
