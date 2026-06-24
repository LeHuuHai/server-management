package model

import "time"

type ServerStatus string

const (
	StatusUnknown ServerStatus = "UNKNOWN"
	StatusOnline  ServerStatus = "ONLINE"
	StatusOffline ServerStatus = "OFFLINE"
)

type LiveStatus struct {
	ServerID        string       `gorm:"primaryKey;not null"`
	Status          ServerStatus `gorm:"type:varchar(20);not null;default:'UNKNOWN'"`
	LastPingAt      time.Time
	LastHeartbeatAt *time.Time // pointer to support nil if never received
}
