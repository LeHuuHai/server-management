package model

import "time"

type ServerUptimeAgg struct {
	ServerID    string
	StartPingAt time.Time
	LastPingAt  time.Time
	UptimeRatio float64
}
