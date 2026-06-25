package model

import (
	"time"

	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type LiveStatus struct {
	ServerID        string                `gorm:"primaryKey;not null"`
	Status          pkgmodel.ServerStatus `gorm:"type:varchar(20);not null;default:'UNKNOWN'"`
	LastPingAt      time.Time
	LastHeartbeatAt *time.Time // pointer to support nil if never received
}
