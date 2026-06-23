package model

import "time"

type ServerStatus string

const (
	StatusUnknown ServerStatus = "UNKNOWN"
	StatusOnline  ServerStatus = "ONLINE"
	StatusOffline ServerStatus = "OFFLINE"
)

type Server struct {
	ServerID          string       `gorm:"primaryKey;not null"`
	ServerName        string       `gorm:"unique;not null"`
	IPv4              string       `gorm:"not null"`
	Status            ServerStatus `gorm:"type:varchar(20);not null;default:'UNKNOWN'"`
	CreatedAt         time.Time    `gorm:"autoCreateTime"`
	MetadataUpdatedAt time.Time    // for name, ip changes
	LastPingAt        time.Time
	Version           int64 `gorm:"default:1"`
	IsDeleted         bool  `gorm:"default:false"`
}
