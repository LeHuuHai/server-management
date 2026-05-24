package model

import "time"

type ResponsePing struct {
	IP     string
	Status string
	PingAt time.Time
}
