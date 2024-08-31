package driver

import (
	"database/sql"
	"log/slog"

	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/optional"
	_ "github.com/lib/pq"
)

type SQL map[string]*sql.DB

func NewSQL(config map[string]config.SQL) (SQL, error) {
	sqls := SQL{}

	for name, dbC := range config {
		if !dbC.Enabled {
			continue
		}

		if dbC.Driver == "postgresql" {
			dbC.Driver = "postgres"
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
			return optional.New(&db)
		}
	}

	if n != "" {
		slog.Info("using default connection", "not_found", n)
	}

	if db, ok := d["default"]; ok {
		return optional.New(&db)
	}

	var db *sql.DB
	return optional.New(&db)
}
