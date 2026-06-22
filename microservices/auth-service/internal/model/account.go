package model

import authdomain "github.com/LeHuuHai/server-management/microservices/pkg/auth"

type Account struct {
	ID       uint            `gorm:"primaryKey"`
	UserID   uint            `gorm:"column:user_id;not null"`
	Username string          `gorm:"column:username;uniqueIndex;not null"`
	Password string          `gorm:"column:password;not null"`
	Role     authdomain.Role `gorm:"column:role;type:varchar(50);not null"`
}

func (a Account) GetUserID() uint {
	return a.UserID
}

func (a Account) GetRole() authdomain.Role {
	return a.Role
}
