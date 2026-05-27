package mq

import "context"

type Consumer interface {
	Read(context.Context) (*Message, error)
	Commit(ctx context.Context, msg *Message) error
}
