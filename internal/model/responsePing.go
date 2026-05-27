package model

import "time"

type ResponsePing struct {
	ServerID string    `json:"server_id"`
	Status   string    `json:"status"`
	PingAt   time.Time `json:"timestamp"`
}
