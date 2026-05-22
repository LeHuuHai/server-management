package kfk

import (
	"log"
	"strings"

	"github.com/LeHuuHai/server-management/config"
	"github.com/segmentio/kafka-go"
)

// return sync writer and async writer
func Connect(config *config.MasterConfig) (*kafka.Writer, *kafka.Writer, error) {
	brokersString := config.Kafka.KafkaBroker
	brokers := strings.Split(brokersString, ",")
	syncWriter := newSyncWriter(brokers)
	asyncWriter := newAsyncWriter(brokers)
	return syncWriter, asyncWriter, nil
}

func newSyncWriter(brokers []string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		RequiredAcks: kafka.RequireOne,
	}
}

func newAsyncWriter(brokers []string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		RequiredAcks: kafka.RequireOne,
		Async:        true,
		Completion: func(messages []kafka.Message, err error) {
			if err != nil {
				log.Printf("send message error: %s", err.Error())
			}
		},
	}
}
