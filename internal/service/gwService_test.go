package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/LeHuuHai/server-management/internal/domain/mq"
	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/LeHuuHai/server-management/internal/service"
)

type mockPublisher struct {
	PublishedMsgs []mq.Message
	PublishFn     func(ctx context.Context, msg mq.Message) error
}

func (m *mockPublisher) Publish(ctx context.Context, msg mq.Message) error {
	if m.PublishFn != nil {
		return m.PublishFn(ctx, msg)
	}
	m.PublishedMsgs = append(m.PublishedMsgs, msg)
	return nil
}

func TestPublishHeartbeat_Success(t *testing.T) {
	publisher := &mockPublisher{}
	svc := service.NewGwService(publisher, "heartbeat-topic")

	hb := model.Heartbeat{
		ServerID:  "srv-001",
		Timestamp: time.Now(),
	}
	err := svc.PublishHeartbeat(context.Background(), hb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(publisher.PublishedMsgs) != 1 {
		t.Fatalf("expected 1 published message, got %d", len(publisher.PublishedMsgs))
	}

	msg := publisher.PublishedMsgs[0]
	if msg.Topic != "heartbeat-topic" {
		t.Errorf("expected topic heartbeat-topic, got %s", msg.Topic)
	}

	var decoded model.Heartbeat
	if err := json.Unmarshal(msg.Value, &decoded); err != nil {
		t.Fatalf("could not decode published message: %v", err)
	}
	if decoded.ServerID != hb.ServerID {
		t.Errorf("expected server_id %s, got %s", hb.ServerID, decoded.ServerID)
	}
}

func TestPublishHeartbeat_PublisherError(t *testing.T) {
	pubErr := errors.New("kafka unavailable")
	hb := model.Heartbeat{ServerID: "s1"}
	publisher := &mockPublisher{
		PublishFn: func(ctx context.Context, msg mq.Message) error {
			if msg.Topic != "heartbeat-topic" {
				t.Errorf("expected topic heartbeat-topic, got %s", msg.Topic)
			}

			var decoded model.Heartbeat
			if err := json.Unmarshal(msg.Value, &decoded); err != nil {
				t.Fatalf("could not decode published message: %v", err)
			}
			if decoded.ServerID != hb.ServerID {
				t.Errorf("expected server_id %s, got %s", hb.ServerID, decoded.ServerID)
			}

			return pubErr
		},
	}
	svc := service.NewGwService(publisher, "heartbeat-topic")
	err := svc.PublishHeartbeat(context.Background(), hb)
	if !errors.Is(err, pubErr) {
		t.Errorf("expected pubErr, got %v", err)
	}
}

func TestPublishHeartbeat_TopicPropagated(t *testing.T) {
	publisher := &mockPublisher{}
	topicName := "custom-heartbeat-topic"
	svc := service.NewGwService(publisher, topicName)

	_ = svc.PublishHeartbeat(context.Background(), model.Heartbeat{ServerID: "x"})

	if len(publisher.PublishedMsgs) != 1 {
		t.Fatal("expected exactly 1 message")
	}
	if publisher.PublishedMsgs[0].Topic != topicName {
		t.Errorf("topic mismatch: expected %s, got %s", topicName, publisher.PublishedMsgs[0].Topic)
	}
}
