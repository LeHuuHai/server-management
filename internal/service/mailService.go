package service

import (
	"context"

	"github.com/LeHuuHai/server-management/internal/domain/mail"
	"github.com/LeHuuHai/server-management/internal/model"
)

type MailService struct {
	sender mail.Sender
}

func (service *MailService) Send(ctx context.Context, mail model.Mail) error {
	return service.sender.Send(ctx, mail)
}

func NewMailService(s mail.Sender) *MailService {
	return &MailService{
		sender: s,
	}
}
