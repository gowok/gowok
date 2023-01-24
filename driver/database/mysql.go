package database

import (
	"github.com/gowok/gowok/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MySQL struct {
	*gorm.DB
}

func NewMysql(conf config.Database) (*MySQL, error) {
	db, err := gorm.Open(mysql.Open(conf.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &MySQL{db}, nil
}
