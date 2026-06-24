package model

import "time"

type StatusLog struct {
	ServerID  string    `json:"server_id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}
