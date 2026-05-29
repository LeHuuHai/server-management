package model

import authdomain "github.com/LeHuuHai/server-management/internal/domain/auth"

type Account struct {
	ID       uint            `gorm:"primaryKey"`
	UserID   uint            `gorm:"column:user_id;not null"`
	UserName string          `gorm:"column:user_name;uniqueIndex;not null"`
	Password string          `gorm:"column:password;not null"`
	Role     authdomain.Role `gorm:"column:role;type:varchar(50);not null"`
}
