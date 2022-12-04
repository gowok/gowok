package database

import "gorm.io/gorm"

type SQLite struct {
	*gorm.DB
}
