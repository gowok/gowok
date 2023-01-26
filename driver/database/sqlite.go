package database

import (
	"database/sql"

	"github.com/gowok/gowok/config"
	_ "github.com/mattn/go-sqlite3"
)

type SQLite struct {
	*sql.DB
}

var _ SQLExecutor = SQLite{}
var _ SQLQuerier = SQLite{}
var _ SQLPreparation = SQLite{}

func NewSqlite(conf config.Database) (*SQLite, error) {
	location := conf.DSN
	if location == "" {
		location = ":memory:"
	}

	db, err := sql.Open("sqlite3", location)
	if err != nil {
		return nil, err
	}

	return &SQLite{db}, nil
}
