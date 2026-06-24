package model

import "time"

type ServerUptimeAgg struct {
	ServerID    string    `json:"server_id"`
	StartPingAt time.Time `json:"start_ping_at"`
	LastPingAt  time.Time `json:"last_ping_at"`
	UptimeRatio float64   `json:"uptime_ratio"`
	DocCount    int64     `json:"doc_count"`
}
