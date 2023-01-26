package database

import (
	"database/sql"

	"github.com/gowok/gowok/config"
	_ "github.com/lib/pq"
)

type PostgreSQL struct {
	*sql.DB
}

var _ SQLExecutor = PostgreSQL{}
var _ SQLQuerier = PostgreSQL{}
var _ SQLPreparation = PostgreSQL{}

func NewPostgresql(conf config.Database) (*PostgreSQL, error) {
	db, err := sql.Open("postgres", conf.DSN)
	if err != nil {
		return nil, err
	}

	return &PostgreSQL{db}, nil
}
