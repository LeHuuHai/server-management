package kfk

import (
	"log"
	"strings"

	commonconfig "github.com/LeHuuHai/server-management/config/common"
	workerconfig "github.com/LeHuuHai/server-management/config/worker"
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

func ConnectWorkerReader(config *workerconfig.KafkaReaderConfig) (*kafka.Reader, *kafka.Reader, error) {
	brokers := strings.Split(config.Broker, ",")
	return kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   config.PingTopic,
			GroupID: config.ConsumerGroupId,
		}),
		kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   config.MailTopic,
			GroupID: config.ConsumerGroupId,
		}),
		nil
}
