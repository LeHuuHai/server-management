package kfk

import (
	"strings"

	"crypto/tls"

	commonconfig "github.com/LeHuuHai/server-management/config/common"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

func newDialer(
	username string,
	password string,
) *kafka.Dialer {
	mechanism := plain.Mechanism{
		Username: username,
		Password: password,
	}

	return &kafka.Dialer{
		SASLMechanism: mechanism,

		// SASL_PLAINTEXT vẫn cần TLS config rỗng
		TLS: &tls.Config{},
	}
}

// return sync writer and async writer
func ConnectWriter(config *commonconfig.KafkaConfig) (*kafka.Writer, *kafka.Writer, error) {
	dialer := newDialer(
		config.Username,
		config.Password,
	)
	brokers := strings.Split(config.Writer.Broker, ",")
	syncWriter := newSyncWriter(brokers, dialer)
	asyncWriter := newAsyncWriter(brokers, dialer)
	return syncWriter, asyncWriter, nil
}

func newSyncWriter(brokers []string, dialer *kafka.Dialer) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		RequiredAcks: kafka.RequireOne,
		Transport: &kafka.Transport{
			SASL: dialer.SASLMechanism,
			TLS:  dialer.TLS,
		},
	}
}

func newAsyncWriter(brokers []string, dialer *kafka.Dialer) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		RequiredAcks: kafka.RequireOne,
		Transport: &kafka.Transport{
			SASL: dialer.SASLMechanism,
			TLS:  dialer.TLS,
		},
	}
}

func ConnectWorkerReader(config *commonconfig.KafkaConfig) (*kafka.Reader, *kafka.Reader, error) {
	dialer := newDialer(
		config.Username,
		config.Password,
	)
	brokers := strings.Split(config.Reader.Broker, ",")
	return kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   config.Topics["ping"],
			GroupID: config.Reader.ConsumerGroupId,
			Dialer:  dialer,
		}),
		kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   config.Topics["mail"],
			GroupID: config.Reader.ConsumerGroupId,
			Dialer:  dialer,
		}),
		nil
}
