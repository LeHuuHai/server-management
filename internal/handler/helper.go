package handler

import (
	"github.com/LeHuuHai/server-management/api"
	"github.com/LeHuuHai/server-management/internal/model"
)

func ServerModelToServerAPI(s model.Server) api.Server {
	return api.Server{
		ServerId:          s.ServerID,
		ServerName:        s.ServerName,
		Status:            api.ServerStatus(s.Status),
		Ipv4:              s.IPv4,
		CreatedAt:         &s.CreatedAt,
		MetadataUpdatedAt: &s.MetadataUpdatedAt,
		LastPingAt:        &s.LastPingAt,
	}
}

func ServerDTOToServerModel(s model.ServerRequestDTO) model.Server {
	return model.Server{
		ServerID:   s.ServerID,
		ServerName: s.ServerName,
		IPv4:       s.IPv4,
	}
}
