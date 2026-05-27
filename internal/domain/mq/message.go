package mq

import "github.com/segmentio/kafka-go"

type Message struct {
	Topic string
	Value []byte
	Raw   *kafka.Message
}
