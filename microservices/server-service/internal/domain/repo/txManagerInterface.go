package repo

import "context"

// TxManagerInterface defines the contract for executing functions within a database transaction.
type TxManagerInterface interface {
	// WithTx executes the provided function within a database transaction.
	// The transaction context is injected into the txCtx.
	WithTx(ctx context.Context, fn func(txCtx context.Context) error) error
}
