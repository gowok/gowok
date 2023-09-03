package gowok

import (
	"log/slog"

	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/optional"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type SQL map[string]*gorm.DB

func NewSQL(config map[string]config.Database) (SQL, error) {
	sqls := SQL{}

	for name, dbC := range config {
		if dbC.Driver == "postgresql" {
			db, err := gorm.Open(postgres.Open(dbC.DSN))
			if err != nil {
				return nil, err
			}

			sqls[name] = db
		}
	}

	return sqls, nil
}

func (d SQL) Get(name ...string) optional.Optional[gorm.DB] {
	n := ""
	if len(name) > 0 {
		n = name[0]
		if db, ok := d[n]; ok {
			return optional.New(db)
		}
	}

	if n != "" {
		slog.Info("using default connection", "not_found", n)
	}

	if db, ok := d["default"]; ok {
		return optional.New(db)
	}

	var db *gorm.DB
	return optional.New(db)
}
