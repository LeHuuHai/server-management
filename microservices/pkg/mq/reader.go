package mq

import (
	"strings"
	"time"

	"github.com/LeHuuHai/server-management/microservices/pkg/config"
	"github.com/segmentio/kafka-go"
)

// ReaderOption defines a functional option for configuring the kafka.Reader
type ReaderOption func(*kafka.ReaderConfig)

// WithMinBytes configures the minimum number of bytes to initiate a fetch
func WithMinBytes(minBytes int) ReaderOption {
	return func(cfg *kafka.ReaderConfig) {
		cfg.MinBytes = minBytes
	}
}

// WithMaxBytes configures the maximum number of bytes to fetch in a single batch
func WithMaxBytes(maxBytes int) ReaderOption {
	return func(cfg *kafka.ReaderConfig) {
		cfg.MaxBytes = maxBytes
	}
}

// WithMaxWait configures the maximum time to wait for a batch to fill before responding
func WithMaxWait(maxWait time.Duration) ReaderOption {
	return func(cfg *kafka.ReaderConfig) {
		cfg.MaxWait = maxWait
	}
}

// WithStartOffset configures the starting offset for the reader (e.g. kafka.FirstOffset, kafka.LastOffset)
func WithStartOffset(offset int64) ReaderOption {
	return func(cfg *kafka.ReaderConfig) {
		cfg.StartOffset = offset
	}
}

// WithCommitInterval configures the interval at which offsets are committed
func WithCommitInterval(interval time.Duration) ReaderOption {
	return func(cfg *kafka.ReaderConfig) {
		cfg.CommitInterval = interval
	}
}

// NewReader creates a new standard Kafka reader using the provided configuration.
// It uses the Functional Options pattern for extension.
func NewReader(cfg *config.KafkaReaderConfig, opts ...ReaderOption) *kafka.Reader {
	rc := kafka.ReaderConfig{
		Brokers:  strings.Split(cfg.Broker, ","),
		Topic:    cfg.Topic,
		GroupID:  cfg.GroupID,
		MinBytes: 10e3, // default 10KB
		MaxBytes: 10e6, // default 10MB
	}

	for _, opt := range opts {
		opt(&rc)
	}

	return kafka.NewReader(rc)
}
