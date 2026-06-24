package model

type RequestPing struct {
	ServerID   string `json:"server_id"`
	ServerName string `json:"server_name"`
	IP         string `json:"ip"`
}
