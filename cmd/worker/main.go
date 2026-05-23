package main

import (
	"context"

	workerconfig "github.com/LeHuuHai/server-management/config/worker"
	kfk "github.com/LeHuuHai/server-management/internal/infra/kafka"
	workerruntime "github.com/LeHuuHai/server-management/internal/infra/runtime/worker"
)

func main() {
	ctx := context.Background()

	cfg, err := workerconfig.Load()
	if err != nil {
		panic(err)
	}

	rt, err := workerruntime.NewApp(cfg)
	if err != nil {
		panic(err)
	}

	// infra
	kfkPublisher := kfk.NewPublisher(rt.SyncWriter)
	kfkConsumer := kfk.NewConsumer(rt.Reader)

}
