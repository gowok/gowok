package database

import "gorm.io/gorm"

type PostgreSQL struct {
	*gorm.DB
}
