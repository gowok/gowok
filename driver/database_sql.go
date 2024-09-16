package driver

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/optional"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type SQL map[string]*sql.DB

func NewSQL(config map[string]config.SQL) (SQL, error) {
	sqls := SQL{}

	for name, dbC := range config {
		if !dbC.Enabled {
			continue
		}

		if dbC.Driver == "postgresql" {
			dbC.Driver = "pgx"
		}

		db, err := sql.Open(dbC.Driver, dbC.DSN)
		if err != nil {
			return nil, err
		}

		sqls[name] = db
	}

	return sqls, nil
}

func (d SQL) Get(name ...string) optional.Optional[*sql.DB] {
	n := ""
	if len(name) > 0 {
		n = name[0]
		if db, ok := d[n]; ok {
			return optional.Of(&db)
		}
	}

	if n != "" {
		slog.Info("using default connection", "not_found", n)
	}

	if db, ok := d["default"]; ok {
		return optional.Of(&db)
	}

	return optional.Empty[*sql.DB]()
}

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
