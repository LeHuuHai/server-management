package mq

import (
	"time"

	"github.com/segmentio/kafka-go"
)

// WriterOption defines a functional option for configuring the kafka.Writer
type WriterOption func(*kafka.Writer)

// WithAsync configures whether the writer should be async
func WithAsync(async bool) WriterOption {
	return func(w *kafka.Writer) {
		w.Async = async
	}
}

// WithBatchSize configures the maximum number of messages sent in a single batch
func WithBatchSize(size int) WriterOption {
	return func(w *kafka.Writer) {
		w.BatchSize = size
	}
}

// WithBatchTimeout configures the maximum time to wait before sending a batch
func WithBatchTimeout(timeout time.Duration) WriterOption {
	return func(w *kafka.Writer) {
		w.BatchTimeout = timeout
	}
}

// WithRequiredAcks configures how many acknowledgements the writer requires
// kafka.RequireNone, kafka.RequireOne, or kafka.RequireAll
func WithRequiredAcks(acks kafka.RequiredAcks) WriterOption {
	return func(w *kafka.Writer) {
		w.RequiredAcks = acks
	}
}

// WithMaxAttempts configures the maximum number of retries
func WithMaxAttempts(attempts int) WriterOption {
	return func(w *kafka.Writer) {
		w.MaxAttempts = attempts
	}
}

// NewWriter creates a new standard Kafka writer for a specific topic.
// It uses the Functional Options pattern so you can add new configs without breaking existing code.
func NewWriter(broker string, topic string, opts ...WriterOption) *kafka.Writer {
	w := &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	// Apply any custom options
	for _, opt := range opts {
		opt(w)
	}

	return w
}
