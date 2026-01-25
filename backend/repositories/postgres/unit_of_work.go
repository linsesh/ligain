package postgres

import (
	"context"
	"database/sql"
)

// txCtxKey is the context key for storing the transaction.
type txCtxKey struct{}

// PostgresUnitOfWork implements UnitOfWork using *sql.DB.
type PostgresUnitOfWork struct {
	db *sql.DB
}

// NewUnitOfWork creates a new PostgresUnitOfWork.
func NewUnitOfWork(db *sql.DB) *PostgresUnitOfWork {
	return &PostgresUnitOfWork{db: db}
}

// WithinTx executes fn within a transaction.
// The transaction is stored in context and can be retrieved by repositories.
func (u *PostgresUnitOfWork) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Put tx in context
	txCtx := context.WithValue(ctx, txCtxKey{}, tx)

	if err := fn(txCtx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// TxFromContext extracts *sql.Tx from context, or returns nil if not present.
func TxFromContext(ctx context.Context) *sql.Tx {
	tx, _ := ctx.Value(txCtxKey{}).(*sql.Tx)
	return tx
}
