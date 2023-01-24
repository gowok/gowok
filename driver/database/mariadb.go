package database

import (
	"github.com/gowok/gowok/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MariaDB struct {
	*gorm.DB
}

func NewMariaDB(conf config.Database) (*MariaDB, error) {
	db, err := gorm.Open(mysql.Open(conf.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &MariaDB{db}, nil
}
