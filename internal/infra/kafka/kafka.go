package kfk

import (
	"log"
	"strings"

	commonconfig "github.com/LeHuuHai/server-management/config/common"
	"github.com/segmentio/kafka-go"
)

// return sync writer and async writer
func ConnectWriter(config *commonconfig.KafkaWriterConfig) (*kafka.Writer, *kafka.Writer, error) {
	brokers := strings.Split(config.Broker, ",")
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

func ConnectReader(config *commonconfig.KafkaReaderConfig) (*kafka.Reader, error) {
	brokers := strings.Split(config.Broker, ",")
	topic := config.Topic
	consumerGroupID := config.ConsumerGroupId
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: consumerGroupID,
	}), nil
}
