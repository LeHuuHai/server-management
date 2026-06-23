package pg

import (
	"context"

	"gorm.io/gorm"
)

type contextKey string

const txKey contextKey = "gorm_tx"

// getDB extracts the transaction from the context if it exists, otherwise returns the default db.
func getDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return tx
	}
	return defaultDB
}

// TxManager handles database transactions.
type TxManager struct {
	db *gorm.DB
}

// NewTxManager creates a new TxManager.
func NewTxManager(db *gorm.DB) *TxManager {
	return &TxManager{db: db}
}

// WithTx starts a transaction and injects it into the context.
func (m *TxManager) WithTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	// If a transaction is already in the context, just reuse it
	if _, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return fn(ctx)
	}

	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		return fn(txCtx)
	})
}
