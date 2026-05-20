package mail

import (
	"context"

	"github.com/LeHuuHai/server-management/internal/model"
)

type Sender interface {
	Send(ctx context.Context, m model.Mail) error
	Close()
}
