package model

import "time"

type ServerRequestDTO struct {
	ServerID   string
	ServerName string
	IPv4       string
}

type ServerResponseDTO struct {
	ServerID    string
	ServerName  string
	IPv4        string
	Status      string
	CreatedTime time.Time
	LastUpdated time.Time
}
