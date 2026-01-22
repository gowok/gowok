package gowok

import (
	"database/sql"
	"log/slog"
	"strings"
	"sync"

	"github.com/gowok/gowok/async"
	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/some"
	"github.com/ngamux/ngamux"
)

var (
	sqlOpen = sql.Open
)

type _sql struct {
	sqls    *sync.Map
	drivers map[string][]string
	plugin  string
}

var SQL = _sql{
	sqls: &sync.Map{},
	drivers: map[string][]string{
		"postgres": {"pgx", "postgres"},
		"mysql":    {"mysql"},
		"mariadb":  {"mysql"},
		"sqlite3":  {"sqlite3"},
	},
	plugin: "sql",
}

func (p *_sql) configure(config map[string]config.SQL) {
	p.sqls = new(sync.Map)
	tasks := make([]func() (any, error), 0)
	for name, dbC := range config {
		if !dbC.Enabled {
			continue
		}
		tasks = append(tasks, func() (any, error) {
			drivers, ok := p.drivers[dbC.Driver]
			if !ok {
				slog.Warn("unknown", "driver", dbC.Driver, "plugin", p.plugin)
				return nil, nil
			}

			for _, driver := range drivers {
				ddb, err := sqlOpen(driver, dbC.DSN)
				if err != nil {
					if strings.Contains(err.Error(), "unknown driver") {
						continue
					}
					slog.Warn("failed to connect", "plugin", p.plugin, "name", name, "error", err)
					return nil, nil
				}

				err = ddb.Ping()
				if err != nil {
					slog.Warn("failed to connect", "plugin", p.plugin, "name", name, "error", err)
					return nil, nil
				}

				p.sqls.Store(name, ddb)

				healthName := "sql"
				if name != "default" {
					healthName += "-" + name
				}
				Health.Add(healthName, p.healthFunc(ddb))
			}

			if _, ok := p.sqls.Load(name); !ok {
				slog.Warn("not installed", "driver", dbC.Driver)
			}

			return nil, nil
		})
	}

	_, err := async.All(tasks...)
	if err != nil {
		slog.Warn("failed to connect", "plugin", p.plugin, "error", err)
		return
	}
}

func (p *_sql) healthFunc(db *sql.DB) func() any {
	return func() any {
		status := ngamux.Map{"status": "DOWN"}
		err := db.Ping()
		if err != nil {
			return status
		}
		return ngamux.Map{"status": "UP"}
	}
}

func (p *_sql) Conn(name ...string) some.Some[*sql.DB] {
	db := p.ConnNoDefault(name...)
	if db.IsPresent() {
		return db
	}

	n := ""
	if len(name) > 0 {
		n = name[0]
	}

	if n == "default" {
		return some.Empty[*sql.DB]()
	}

	if n != "" {
		slog.Info("using default connection", "not_found", n)
	}

	return p.Conn("default")
}

func (p *_sql) ConnNoDefault(name ...string) some.Some[*sql.DB] {
	n := ""
	if len(name) > 0 {
		n = name[0]
		if db, ok := p.sqls.Load(n); ok {
			if db, ok := db.(*sql.DB); ok {
				return some.Of(db)
			}
		}
	}

	return some.Empty[*sql.DB]()
}
