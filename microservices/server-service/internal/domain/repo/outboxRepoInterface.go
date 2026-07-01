package repo

import (
	"context"

	"github.com/LeHuuHai/server-management/microservices/server-service/internal/model"
)

// OutboxRepositoryInterface defines the contract for outbox operations.
type OutboxRepositoryInterface interface {
	// CreateEvent creates a new outbox event, typically within a transaction.
	CreateEvent(ctx context.Context, event *model.OutboxEvent) error

	// FetchPendingEvents retrieves up to 'limit' pending outbox events.
	FetchPendingEvents(ctx context.Context, limit int) ([]model.OutboxEvent, error)

	// DeleteEvents deletes outbox events by their IDs after successful publishing.
	DeleteEvents(ctx context.Context, ids []string) error

	// MarkEventsDone marks outbox events as done after successful publishing.
	MarkEventsDone(ctx context.Context, ids []string) error
}
