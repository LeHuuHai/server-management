package model

type ServerStatus string

const (
	StatusUnknown ServerStatus = "UNKNOWN"
	StatusOnline  ServerStatus = "ONLINE"
	StatusOffline ServerStatus = "OFFLINE"
)
