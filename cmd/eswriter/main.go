package main

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	eswriterconfig "github.com/LeHuuHai/server-management/config/eswriter"
	es "github.com/LeHuuHai/server-management/internal/infra/elasticsearch"
	kfk "github.com/LeHuuHai/server-management/internal/infra/kafka"
	eswriterruntime "github.com/LeHuuHai/server-management/internal/infra/runtime/eswriter"
	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/LeHuuHai/server-management/internal/service"
)

func ReadPingResTopic(
	ctx context.Context,
	wg *sync.WaitGroup,
	consumer *kfk.KfkConsumer,
	ch chan<- model.ServerEvent,
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
		consumer.Commit(ctx, msg)
		var res model.ResponsePing
		err = json.Unmarshal(msg.Value, &res)
		if err != nil {
			continue
		}
		serverEvent := model.ServerEvent{
			ServerID:  res.ServerID,
			Status:    res.Status,
			Timestamp: time.Now().UTC(),
		}
		select {
		case <-ctx.Done():
			return

		case ch <- serverEvent:
		}
	}
}

func ReadHeartbeatTopic(
	ctx context.Context,
	wg *sync.WaitGroup,
	consumer *kfk.KfkConsumer,
	ch chan<- model.ServerEvent,
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
		consumer.Commit(ctx, msg)
		var res model.Heartbeat
		err = json.Unmarshal(msg.Value, &res)
		if err != nil {
			continue
		}
		serverEvent := model.ServerEvent{
			ServerID:  res.ServerID,
			Status:    "on",
			Timestamp: res.Timestamp,
		}
		select {
		case <-ctx.Done():
			return

		case ch <- serverEvent:
		}
	}
}

func main() {
	ctx := context.Background()

	cfg, err := eswriterconfig.Load()
	if err != nil {
		panic(err)
	}

	rt, err := eswriterruntime.NewApp(cfg)
	if err != nil {
		panic(err)
	}

	// domain, infra
	pingResConsumer := kfk.NewConsumer(rt.PingResReader)
	heartbeatConsumer := kfk.NewConsumer(rt.HeartbeatReader)
	writer := es.NewWriter[model.ServerEvent](rt.ESClient, rt.Config.ESConfig.Index)

	// service
	ch := make(chan model.ServerEvent, 4000)
	batchService := service.NewBatchESService(
		ch,
		2000,
		time.Second,
		func(items []model.ServerEvent) error {
			return writer.WriteBatch(items)
		},
	)

	var wg sync.WaitGroup
	wg.Add(3)
	go ReadPingResTopic(ctx, &wg, pingResConsumer, ch)
	go ReadHeartbeatTopic(ctx, &wg, heartbeatConsumer, ch)
	go func() {
		defer wg.Done()
		batchService.Run(ctx)
	}()
	wg.Wait()
}
