package kfk

import (
	"strings"

	commonconfig "github.com/LeHuuHai/server-management/config/common"
	"github.com/segmentio/kafka-go"
)

// return sync writer and async writer
func ConnectWriter(config *commonconfig.KafkaConfig) (*kafka.Writer, *kafka.Writer, error) {
	brokers := strings.Split(config.Writer.Broker, ",")
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
	}
}

func ConnectWorkerReader(config *commonconfig.KafkaConfig) (*kafka.Reader, *kafka.Reader, error) {
	brokers := strings.Split(config.Reader.Broker, ",")
	return kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   config.Topics["ping"],
			GroupID: config.Reader.ConsumerGroupId,
		}),
		kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   config.Topics["mail"],
			GroupID: config.Reader.ConsumerGroupId,
		}),
		nil
}
