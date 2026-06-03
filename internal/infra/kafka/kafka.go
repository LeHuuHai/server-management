package kfk

import (
	"log/slog"
	"strings"
	"time"

	commonconfig "github.com/LeHuuHai/server-management/config/common"
	"github.com/segmentio/kafka-go"
)

// return sync writer and async writer
func ConnectWriter(config *commonconfig.KafkaConfig) (*kafka.Writer, *kafka.Writer, error) {
	brokers := strings.Split(config.Writer.Broker, ",")
	syncWriter := newSyncWriter(brokers)
	asyncWriter := newAsyncWriter(brokers)
	slog.Info("Kafka writers connected", "brokers", brokers)
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
		Async:        true,
		BatchSize:    1000,
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
	}
}

func ConnectWorkerReader(config *commonconfig.KafkaConfig) (*kafka.Reader, *kafka.Reader, error) {
	brokers := strings.Split(config.Reader.Broker, ",")
	pingReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       config.Topics["ping"],
		GroupID:     config.Reader.ConsumerGroupId,
		StartOffset: kafka.LastOffset,
	})
	slog.Info("Kafka ping readers connected", "brokers", brokers, "topics", config.Topics["ping"], "groupId", config.Reader.ConsumerGroupId)
	mailReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       config.Topics["mail"],
		GroupID:     config.Reader.ConsumerGroupId,
		StartOffset: kafka.LastOffset,
	})
	slog.Info("Kafka mail readers connected", "brokers", brokers, "topics", config.Topics["mail"], "groupId", config.Reader.ConsumerGroupId)
	return pingReader, mailReader, nil
}

func ConnectPingResReader(config *commonconfig.KafkaConfig) (*kafka.Reader, error) {
	brokers := strings.Split(config.Reader.Broker, ",")
	pingResReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       config.Topics["ping_res"],
		GroupID:     config.Reader.ConsumerGroupId,
		StartOffset: kafka.LastOffset,
	})
	slog.Info("Kafka ping res readers connected", "brokers", brokers, "topics", config.Topics["ping_res"], "groupId", config.Reader.ConsumerGroupId)
	return pingResReader, nil
}

func ConnectHeartbeatReader(config *commonconfig.KafkaConfig) (*kafka.Reader, error) {
	brokers := strings.Split(config.Reader.Broker, ",")
	heartbeatReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       config.Topics["heartbeat"],
		GroupID:     config.Reader.ConsumerGroupId,
		StartOffset: kafka.LastOffset,
	})
	slog.Info("Kafka heartbeat readers connected", "brokers", brokers, "topics", config.Topics["heartbeat"], "groupId", config.Reader.ConsumerGroupId)
	return heartbeatReader, nil
}
