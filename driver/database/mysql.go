package database

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gowok/gowok/config"
)

type MySQL struct {
	*sql.DB
}

var _ SQLExecutor = MySQL{}
var _ SQLQuerier = MySQL{}
var _ SQLPreparation = MySQL{}

func NewMysql(conf config.Database) (*MySQL, error) {
	db, err := sql.Open("mysql", conf.DSN)
	if err != nil {
		return nil, err
	}

	return &MySQL{db}, nil
}
