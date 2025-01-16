package sql

import (
	"database/sql"
	"log/slog"
	"strings"
	"sync"

	"github.com/gowok/gowok/async"
	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/health"
	"github.com/gowok/gowok/some"
	"github.com/ngamux/ngamux"
)

var plugin = "sql"
var sqls *sync.Map
var drivers = map[string][]string{
	"postgres": []string{"pgx", "postgres"},
	"mysql":    []string{"mysql"},
	"mariadb":  []string{"mysql"},
	"sqlite3":  []string{"sqlite3"},
}

func Configure(config map[string]config.SQL) {
	sqls = new(sync.Map)
	tasks := make([]func() (any, error), 0)
	for name, dbC := range config {
		if !dbC.Enabled {
			continue
		}
		tasks = append(tasks, func() (any, error) {
			drivers, ok := drivers[dbC.Driver]
			if !ok {
				slog.Warn("unknown", "driver", dbC.Driver, "plugin", plugin)
				return nil, nil
			}

			for _, driver := range drivers {
				ddb, err := sql.Open(driver, dbC.DSN)
				if err != nil {
					if strings.Contains(err.Error(), "unknown driver") {
						continue
					}
					slog.Warn("failed to connect", "plugin", plugin, "name", name, "error", err)
					return nil, nil
				}

				err = ddb.Ping()
				if err != nil {
					slog.Warn("failed to connect", "plugin", plugin, "name", name, "error", err)
					return nil, nil
				}

				sqls.Store(name, ddb)

				healthName := "sql"
				if name != "default" {
					healthName += "-" + name
				}
				health.Add(healthName, healthFunc(ddb))
			}

			if _, ok := sqls.Load(name); !ok {
				slog.Warn("not installed", "driver", dbC.Driver)
			}

			return nil, nil
		})
	}

	_, err := async.All(tasks...)
	if err != nil {
		slog.Warn("failed to connect", "plugin", plugin, "error", err)
		return
	}
}

func DB(name ...string) some.Some[*sql.DB] {
	n := ""
	if len(name) > 0 {
		n = name[0]
		if db, ok := sqls.Load(n); ok {
			if db, ok := db.(*sql.DB); ok {
				return some.Of(db)
			}
		}
	}

	if n != "" {
		slog.Info("using default connection", "not_found", n)
	}

	if db, ok := sqls.Load("default"); ok {
		if db, ok := db.(*sql.DB); ok {
			return some.Of(db)
		}
	}

	return some.Empty[*sql.DB]()
}

func DBNoDefault(name ...string) some.Some[*sql.DB] {
	n := ""
	if len(name) > 0 {
		n = name[0]
		if db, ok := sqls.Load(n); ok {
			if db, ok := db.(*sql.DB); ok {
				return some.Of(db)
			}
		}
	}

	return some.Empty[*sql.DB]()
}

func healthFunc(db *sql.DB) func() any {
	return func() any {
		status := ngamux.Map{"status": "DOWN"}
		err := db.Ping()
		if err != nil {
			return status
		}
		return ngamux.Map{"status": "UP"}
	}
}
