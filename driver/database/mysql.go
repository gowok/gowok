package database

import "gorm.io/gorm"

type MySQL struct {
	*gorm.DB
}
