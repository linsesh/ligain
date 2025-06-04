package postgres

import (
	"context"
	"database/sql"
)

// DBExecutor interface allows using either *sql.DB or *sql.Tx
// This enables dependency injection for better testing and transaction support
type DBExecutor interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type PostgresRepository struct {
	db DBExecutor
}

func NewPostgresRepository(db DBExecutor) *PostgresRepository {
	return &PostgresRepository{db: db}
}
