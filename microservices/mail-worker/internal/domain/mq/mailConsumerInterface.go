package mq

import (
	"context"

	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type MailConsumerInterface interface {
	Read(ctx context.Context) (pkgmodel.RequestMail, func(context.Context) error, error)
	Close() error
}
