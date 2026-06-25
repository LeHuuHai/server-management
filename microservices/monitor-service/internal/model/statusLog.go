package model

import (
	"time"

	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type StatusLog struct {
	ServerID  string                `json:"server_id"`
	Status    pkgmodel.ServerStatus `json:"status"`
	Timestamp time.Time             `json:"timestamp"`
}
