package model

import "time"

type ResponsePing struct {
	ServerID string
	Status   string
	PingAt   time.Time
}
