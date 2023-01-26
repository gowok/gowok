package database

import (
	"database/sql"

	"github.com/gowok/gowok/config"
)

type MariaDB struct {
	*sql.DB
}

var _ SQLExecutor = MariaDB{}
var _ SQLQuerier = MariaDB{}
var _ SQLPreparation = MariaDB{}

func NewMariaDB(conf config.Database) (*MariaDB, error) {
	db, err := sql.Open("mysql", conf.DSN)
	if err != nil {
		return nil, err
	}

	return &MariaDB{db}, nil
}
