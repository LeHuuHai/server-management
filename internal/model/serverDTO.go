package model

import "time"

type ServerRequestDTO struct {
	ServerID   string `json:"server_id"`
	ServerName string `json:"server_name"`
	IPv4       string `json:"ipv4"`
}

type ServerResponseDTO struct {
	ServerID          string
	ServerName        string
	IPv4              string
	Status            string
	CreatedTime       time.Time
	MetadataUpdatedAt time.Time
	LastPingAt        time.Time
}

type ServerImport struct {
	ServerID   string
	ServerName string
	IPv4       string
}
