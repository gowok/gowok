package database

import "gorm.io/gorm"

type MariaDB struct {
	*gorm.DB
}
