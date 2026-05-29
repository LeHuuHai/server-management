package model

type User struct {
	ID       uint   `gorm:"primaryKey"`
	FullName string `gorm:"column:full_name;not null"`
	Email    string `gorm:"column:email;uniqueIndex;not null"`
}
