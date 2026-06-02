package model

import "time"

type ServerMetadata struct {
	ServerID        string
	ServerName      string
	IPv4            string
	LastHeartbeatAt *time.Time // nil nếu chưa bao giờ gửi heartbeat
}
