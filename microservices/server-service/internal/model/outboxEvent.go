package model

import "time"

const (
	OutboxStatusPending = "PENDING"
	OutboxStatusDone    = "DONE"
)

// OutboxEvent represents an event that needs to be published to a message broker.
type OutboxEvent struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Topic     string    `gorm:"type:varchar(255);not null" json:"topic"`
	Payload   []byte    `gorm:"type:jsonb;not null" json:"payload"`
	Status    string    `gorm:"type:varchar(50);not null;default:'PENDING'" json:"status"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}
