package mq

type Message struct {
	Topic string
	Value []byte
}
