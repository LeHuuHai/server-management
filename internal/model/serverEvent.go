package model

import "time"

type ServerEvent struct {
	ServerID  string    `json:"server_id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}
