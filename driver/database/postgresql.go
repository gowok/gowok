package database

import (
	"github.com/gowok/gowok/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgreSQL struct {
	*gorm.DB
}

func configurePostgresql(conf config.Database) (*PostgreSQL, error) {
	db, err := gorm.Open(postgres.Open(conf.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &PostgreSQL{db}, nil
}
