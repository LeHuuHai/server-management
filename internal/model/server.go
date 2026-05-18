package model

import "time"

type Server struct {
	ServerID    string    `gorm:"primaryKey;not null"`
	ServerName  string    `gorm:"unique;not null"`
	IPv4        string    `gorm:"unique;not null"`
	Status      string    `gorm:"not null"`
	CreatedTime time.Time `gorm:"autoCreateTime"`
	LastUpdated time.Time `gorm:"autoUpdateTime"`
	IsDeleted   bool      `gorm:"default:false"`
}
