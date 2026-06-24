package model

type MonitoredServer struct {
	ServerID   string `gorm:"primaryKey;not null"`
	ServerName string `gorm:"unique;not null"`
	IPv4       string `gorm:"not null"`
	Version    int64  `gorm:"default:1"`
}
