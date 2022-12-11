package database

import (
	"context"
	"database/sql"
)

type SQLPreparation interface {
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

type SQLQuerier interface {
	PingContext(ctx context.Context) error
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type SQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type SQLTx interface {
	SQLQuerier
	SQLExecutor
	Commit() error
	Rollback() error
}
