package repositories

import "context"

// UnitOfWork defines atomic transaction boundaries.
// The infrastructure layer tucks the transaction into context.
type UnitOfWork interface {
	// WithinTx executes fn within a transaction.
	// If fn returns an error, the transaction is rolled back.
	// If fn returns nil, the transaction is committed.
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
