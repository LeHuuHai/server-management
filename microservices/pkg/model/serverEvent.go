package model

import "time"

type ServerEvent struct {
	ServerID   string    `json:"server_id"`
	ServerName string    `json:"server_name"`
	IPv4       string    `json:"ipv4"`
	Timestamp  time.Time `json:"timestamp"`
	Version    int64     `json:"version"`
}
