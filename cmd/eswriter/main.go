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

func ReadTopic(
	ctx context.Context,
	wg *sync.WaitGroup,
	consumer *kfk.KfkConsumer,
	ch chan<- model.ResponsePing,
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
		select {
		case <-ctx.Done():
			return

		case ch <- res:
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
	consumer := kfk.NewConsumer(rt.HeartBeatReader)
	writer := es.NewWriter[model.ResponsePing](rt.ESClient, rt.Config.ESConfig.Index)

	// service
	ch := make(chan model.ResponsePing, 10000)
	batchService := service.NewBatchService(
		ch,
		2000,
		time.Second,
		func(items []model.ResponsePing) error {
			return writer.WriteBatch(items)
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
