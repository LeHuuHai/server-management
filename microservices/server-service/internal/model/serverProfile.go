package model

import "time"

type ServerProfile struct {
	ServerID   string    `gorm:"primaryKey;not null"`
	ServerName string    `gorm:"unique;not null"`
	IPv4       string    `gorm:"not null"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
	Version    int64     `gorm:"default:1"`
	IsDeleted  bool      `gorm:"default:false"`
}
