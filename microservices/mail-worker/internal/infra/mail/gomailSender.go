package mail

import (
	"bytes"
	"context"
	"io"

	"github.com/LeHuuHai/server-management/microservices/mail-worker/internal/domain/mail"
	pkgmodel "github.com/LeHuuHai/server-management/microservices/pkg/model"
	"gopkg.in/gomail.v2"
)

type GomailSender struct {
	dialer   *gomail.Dialer
	fromAddr string
}

func NewGomailSender(dialer *gomail.Dialer, fromAddr string) mail.SenderInterface {
	return &GomailSender{
		dialer:   dialer,
		fromAddr: fromAddr,
	}
}

func (s *GomailSender) Send(ctx context.Context, mailPayload pkgmodel.Mail) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.fromAddr)
	m.SetHeader("To", mailPayload.To...)
	m.SetHeader("Subject", mailPayload.Subject)
	m.SetBody("text/html", mailPayload.Body)

	for _, attachment := range mailPayload.Attachments {
		m.Attach(attachment.Filename, gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := io.Copy(w, bytes.NewReader(attachment.Data))
			return err
		}))
	}

	return s.dialer.DialAndSend(m)
}
