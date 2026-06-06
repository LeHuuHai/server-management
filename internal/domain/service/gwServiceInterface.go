package serviceinterface

import (
	"context"

	"github.com/LeHuuHai/server-management/internal/model"
)

type GWServiceInterface interface {
	PublishHeartbeat(ctx context.Context, heartbeat model.Heartbeat) error
}
