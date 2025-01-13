package sql

import (
	"database/sql"
	"log/slog"
	"strings"

	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/some"
)

var plugin = "sql"
var sqls map[string]*sql.DB
var drivers = map[string][]string{
	"postgres": []string{"pgx", "postgres"},
	"mysql":    []string{"mysql"},
	"mariadb":  []string{"mysql"},
	"sqlite3":  []string{"sqlite3"},
}

func Configure(config map[string]config.SQL) {
	sqls = make(map[string]*sql.DB)
	for name, dbC := range config {
		if !dbC.Enabled {
			continue
		}

		drivers, ok := drivers[dbC.Driver]
		if !ok {
			slog.Warn("unknown", "driver", dbC.Driver, "plugin", plugin)
			continue
		}

		for _, driver := range drivers {
			ddb, err := sql.Open(driver, dbC.DSN)
			if err != nil {
				if strings.Contains(err.Error(), "unknown driver") {
					continue
				}
				slog.Warn("failed to connect", "plugin", plugin, "name", name, "error", err)
				return
			}

			err = ddb.Ping()
			if err != nil {
				slog.Warn("failed to connect", "plugin", plugin, "name", name, "error", err)
				return
			}

			sqls[name] = ddb
		}

		if _, ok := sqls[name]; !ok {
			slog.Warn("not installed", "driver", dbC.Driver)
		}
	}
}

func DB(name ...string) some.Some[*sql.DB] {
	n := ""
	if len(name) > 0 {
		n = name[0]
		if db, ok := sqls[n]; ok {
			return some.Of(db)
		}
	}

	if n != "" {
		slog.Info("using default connection", "not_found", n)
	}

	if db, ok := sqls["default"]; ok {
		return some.Of(db)
	}

	return some.Empty[*sql.DB]()
}

func GetNoDefault(name ...string) some.Some[*sql.DB] {
	n := ""
	if len(name) > 0 {
		n = name[0]
		if db, ok := sqls[n]; ok {
			return some.Of(db)
		}
	}

	return some.Empty[*sql.DB]()
}

func Ping() map[string]string {
  
  	var result = make(map[string]string)

  	for name, dbConn := range sqls {
		if err := dbConn.Ping(); err != nil {
			result[name] = "un-healty"
		}else {
			result[name] = "healty"
		}
  	}

  	return result
}
