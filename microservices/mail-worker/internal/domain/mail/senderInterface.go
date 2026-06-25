package mail

import (
	"context"

	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type SenderInterface interface {
	Send(ctx context.Context, mail pkgmodel.Mail) error
}
