package base

import "gorm.io/gorm"

type Model struct {
	DB *gorm.DB
}
