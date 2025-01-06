package driver

import (
	"context"
	"database/sql"
	"log/slog"
	"strings"

	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/some"
)

type SQL map[string]*sql.DB

var drivers = map[string][]string{
	"postgres": []string{"pgx", "postgres"},
	"mysql":    []string{"mysql"},
	"mariadb":  []string{"mysql"},
	"sqlite3":  []string{"sqlite3"},
}

func NewSQL(config map[string]config.SQL) (SQL, error) {
	sqls := SQL{}

	for name, dbC := range config {
		if !dbC.Enabled {
			continue
		}

		drivers, ok := drivers[dbC.Driver]
		if !ok {
			slog.Warn("unknown SQL", "driver", dbC.Driver)
			continue
		}

		for _, driver := range drivers {
			ddb, err := sql.Open(driver, dbC.DSN)
			if err != nil {
				if strings.Contains(err.Error(), "unknown driver") {
					continue
				}
				return nil, err
			}

			sqls[name] = ddb
		}

		if _, ok := sqls[name]; !ok {
			slog.Warn("not installed", "driver", dbC.Driver)
		}
	}

	return sqls, nil
}

func (d SQL) Get(name ...string) some.Some[*sql.DB] {
	n := ""
	if len(name) > 0 {
		n = name[0]
		if db, ok := d[n]; ok {
			return some.Of(db)
		}
	}

	if n != "" {
		slog.Info("using default connection", "not_found", n)
	}

	if db, ok := d["default"]; ok {
		return some.Of(db)
	}

	return some.Empty[*sql.DB]()
}

func (d SQL) GetNoDefault(name ...string) some.Some[*sql.DB] {
	n := ""
	if len(name) > 0 {
		n = name[0]
		if db, ok := d[n]; ok {
			return some.Of(db)
		}
	}

	return some.Empty[*sql.DB]()
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
