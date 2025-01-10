package sql

import (
	"context"
	"database/sql"
)

type Preparation interface {
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

type Querier interface {
	PingContext(ctx context.Context) error
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type Tx interface {
	Querier
	Executor
	Commit() error
	Rollback() error
}
