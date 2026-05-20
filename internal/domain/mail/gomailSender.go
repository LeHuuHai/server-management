package mail

import (
	"context"
	"sync"

	"github.com/LeHuuHai/server-management/internal/model"
	"gopkg.in/gomail.v2"
)

type GomailSender struct {
	dialer *gomail.Dialer
	conn   gomail.SendCloser
	mu     sync.Mutex
}

func NewGomailSender(d *gomail.Dialer) (*GomailSender, error) {
	conn, err := d.Dial()
	if err != nil {
		return nil, err
	}
	return &GomailSender{
		dialer: d,
		conn:   conn,
	}, nil
}

func (s *GomailSender) Send(ctx context.Context, mail model.Mail) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", mail.From)
	msg.SetHeader("To", mail.To...)
	msg.SetHeader("Subject", mail.Subject)
	msg.SetBody("text/plain", mail.Body)
	for _, item := range mail.Attachments {
		msg.Attach(item.Path, gomail.Rename(item.Filename))
	}
	return s.sendWithRetry(msg)
}

func (s *GomailSender) sendWithRetry(m *gomail.Message) error {
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

func (s *GomailSender) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
