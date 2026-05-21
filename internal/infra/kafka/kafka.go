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
	topic := config.Kafka.KafkaTopic
	syncWriter := newSyncWriter(brokers, topic)
	asyncWriter := newAsyncWriter(brokers, topic)
	return syncWriter, asyncWriter, nil
}

func newSyncWriter(brokers []string, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		RequiredAcks: kafka.RequireOne,
	}
}

func newAsyncWriter(brokers []string, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		RequiredAcks: kafka.RequireOne,
		Async:        true,
		Completion: func(messages []kafka.Message, err error) {
			if err != nil {
				log.Printf("send message error: %s", err.Error())
			}
		},
	}
}
