package service

import (
	"context"
	"time"

	"github.com/LeHuuHai/server-management/microservices/ping-worker/internal/domain/service"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
	probing "github.com/prometheus-community/pro-bing"
)

type PingService struct{}

func NewPingService() service.PingServiceInterface {
	return &PingService{}
}

func (s *PingService) Ping(ctx context.Context, req pkgmodel.RequestPing) (pkgmodel.ResponsePing, error) {
	res := pkgmodel.ResponsePing{
		ServerID: req.ServerID,
		Status:   pkgmodel.StatusOffline,
		PingAt:   time.Now(),
	}

	pinger, err := probing.NewPinger(req.IP)
	if err == nil {
		pinger.Count = 1
		pinger.Timeout = 1 * time.Second
		pinger.SetPrivileged(true)

		err = pinger.Run()
		if err == nil {
			stats := pinger.Statistics()
			if stats.PacketsRecv > 0 {
				res.Status = pkgmodel.StatusOnline
			}
		}
	}

	return res, nil
}
