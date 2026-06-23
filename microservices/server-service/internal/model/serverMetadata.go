package model

import "time"

type ServerMetadata struct {
	ServerID        string
	ServerName      string
	IPv4            string
	LastHeartbeatAt *time.Time // nil náº¿u chÆ°a bao giá» gá»­i heartbeat
}
