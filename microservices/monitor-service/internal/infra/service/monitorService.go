package service

import (
	"context"

	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/repo"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/domain/service"
	"github.com/LeHuuHai/server-management/microservices/monitor-service/internal/model"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type MonitorService struct {
	monitoredServerRepo repo.MonitoredServerRepositoryInterface
	liveStatusRepo      repo.LiveStatusRepositoryInterface
	pgChan              chan model.LiveStatus
	esChan              chan model.StatusLog
}

func NewMonitorService(
	monitoredServerRepo repo.MonitoredServerRepositoryInterface,
	liveStatusRepo repo.LiveStatusRepositoryInterface,
	pgChan chan model.LiveStatus,
	esChan chan model.StatusLog,
) service.MonitorServiceInterface {
	return &MonitorService{
		monitoredServerRepo: monitoredServerRepo,
		liveStatusRepo:      liveStatusRepo,
		pgChan:              pgChan,
		esChan:              esChan,
	}
}

func (s *MonitorService) ProcessHeartbeat(ctx context.Context, hb pkgmodel.Heartbeat) error {
	server := model.LiveStatus{
		ServerID:        hb.ServerID,
		Status:          pkgmodel.StatusOnline,
		LastHeartbeatAt: &hb.Timestamp,
	}
	s.pgChan <- server

	event := model.StatusLog{
		ServerID:  hb.ServerID,
		Status:    pkgmodel.StatusOnline,
		Timestamp: hb.Timestamp,
	}
	s.esChan <- event

	return nil
}

func (s *MonitorService) ProcessPingResult(ctx context.Context, res pkgmodel.ResponsePing) error {
	statusVal := pkgmodel.StatusOffline
	if res.Status == string(pkgmodel.StatusOnline) {
		statusVal = pkgmodel.StatusOnline
	}

	server := model.LiveStatus{
		ServerID:   res.ServerID,
		Status:     statusVal,
		LastPingAt: res.PingAt,
	}
	s.pgChan <- server

	event := model.StatusLog{
		ServerID:  res.ServerID,
		Status:    statusVal,
		Timestamp: res.PingAt,
	}
	s.esChan <- event

	return nil
}

func (s *MonitorService) SyncServerLifecycle(ctx context.Context, event pkgmodel.ServerEvent, action string) error {
	switch action {
	case "Create":
		srv := &model.MonitoredServer{
			ServerID:   event.ServerID,
			ServerName: event.ServerName,
			IPv4:       event.IPv4,
			Version:    event.Version,
		}
		if err := s.monitoredServerRepo.Create(ctx, srv); err != nil {
			return err
		}

		status := &model.LiveStatus{
			ServerID: event.ServerID,
			Status:   pkgmodel.StatusUnknown,
		}
		return s.liveStatusRepo.Create(ctx, status)

	case "Update":
		srv := &model.MonitoredServer{
			ServerID:   event.ServerID,
			ServerName: event.ServerName,
			IPv4:       event.IPv4,
			Version:    event.Version,
		}
		return s.monitoredServerRepo.Update(ctx, srv)

	case "Delete":
		if err := s.monitoredServerRepo.Delete(ctx, event.ServerID); err != nil {
			return err
		}
		return s.liveStatusRepo.Delete(ctx, event.ServerID)
	}
	return nil
}
