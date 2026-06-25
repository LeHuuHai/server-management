package model

import "time"

type Heartbeat struct {
	ServerID  string    `json:"server_id"`
	Timestamp time.Time `json:"timestamp"`
}
