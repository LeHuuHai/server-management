package mq

import (
	"context"

	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
)

type MailPublisherInterface interface {
	Publish(ctx context.Context, req pkgmodel.RequestMail) error
}
