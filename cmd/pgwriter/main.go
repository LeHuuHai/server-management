package main

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	pgwriterconfig "github.com/LeHuuHai/server-management/config/pgwriter"
	kfk "github.com/LeHuuHai/server-management/internal/infra/kafka"
	pg "github.com/LeHuuHai/server-management/internal/infra/postgres"
	pgwriterruntime "github.com/LeHuuHai/server-management/internal/infra/runtime/pgwriter"
	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/LeHuuHai/server-management/internal/service"
)

func ReadTopic(
	ctx context.Context,
	wg *sync.WaitGroup,
	consumer *kfk.KfkConsumer,
	ch chan<- model.Server,
) {
	defer wg.Done()
	for {
		msg, err := consumer.Read(ctx)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				continue
			}
		}
		var res model.ResponsePing
		err = json.Unmarshal(msg.Value, &res)
		if err != nil {
			continue
		}
		s := model.Server{
			ServerID:   res.ServerID,
			Status:     model.ServerStatus(res.Status),
			LastPingAt: res.PingAt,
		}
		select {
		case <-ctx.Done():
			return

		case ch <- s:
			consumer.Commit(ctx, msg)
		}
	}
}

func main() {
	ctx := context.Background()

	cfg, err := pgwriterconfig.Load()
	if err != nil {
		panic(err)
	}

	rt, err := pgwriterruntime.NewApp(cfg)
	if err != nil {
		panic(err)
	}

	// domain, infra
	consumer := kfk.NewConsumer(rt.HeartBeatReader)
	serverRepo := pg.NewServerRepository(rt.DB)

	// service
	ch := make(chan model.Server, 2000)
	batchService := service.NewBatchPGService(
		ch,
		1000,
		time.Second,
		func(items map[string]model.Server) error {
			if len(items) == 0 {
				return nil
			}
			values := make([]model.Server, 0, len(items))
			for _, v := range items {
				values = append(values, v)
			}
			return serverRepo.BulkUpdateServers(ctx, values)
		},
	)

	var wg sync.WaitGroup
	wg.Add(2)
	go ReadTopic(ctx, &wg, consumer, ch)
	go func() {
		defer wg.Done()
		batchService.Run(ctx)
	}()
	wg.Wait()
}
