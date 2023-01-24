package database

import (
	"github.com/gowok/gowok/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SQLite struct {
	*gorm.DB
}

func NewSqlite(conf config.Database) (*SQLite, error) {
	location := conf.DSN
	if location == "" {
		location = "db.sqlite3"
	}

	db, err := gorm.Open(sqlite.Open(location), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &SQLite{db}, nil
}
