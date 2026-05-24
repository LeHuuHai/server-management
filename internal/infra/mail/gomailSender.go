package smtp

import (
	"context"
	"sync"

	"github.com/LeHuuHai/server-management/internal/model"
	"gopkg.in/gomail.v2"
)

type gomailSender struct {
	dialer *gomail.Dialer
	conn   gomail.SendCloser
	mu     sync.Mutex
	From   string
}

func NewGomailSender(d *gomail.Dialer, from string) (*gomailSender, error) {
	conn, err := d.Dial()
	if err != nil {
		return nil, err
	}
	return &gomailSender{
		dialer: d,
		conn:   conn,
		From:   from,
	}, nil
}

func (s *gomailSender) Send(ctx context.Context, mail model.Mail) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", s.From)
	msg.SetHeader("To", mail.To...)
	msg.SetHeader("Subject", mail.Subject)
	msg.SetBody("text/plain", mail.Body)
	for _, item := range mail.Attachments {
		msg.Attach(item.Path, gomail.Rename(item.Filename))
	}
	return s.sendWithRetry(msg)
}

func (s *gomailSender) sendWithRetry(m *gomail.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// try send first time
	if err := gomail.Send(s.conn, m); err == nil {
		return nil
	}

	// reconnect
	conn, err := s.dialer.Dial()
	if err != nil {
		return err
	}
	_ = s.conn.Close()
	s.conn = conn

	// retry once
	return gomail.Send(s.conn, m)
}

func (s *gomailSender) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil {
		s.conn.Close()
	}
}
