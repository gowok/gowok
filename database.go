package gowok

import (
	"github.com/gowok/gowok/base"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func (app *App) dbConnect() {
	var err error
	if app.Config.DB.Driver == base.DriverMySQL {
		app.db, err = gorm.Open(mysql.Open(app.Config.DB.DSN()))
	} else if app.Config.DB.Driver == base.DriverPostgreSQL {
		app.db, err = gorm.Open(postgres.Open(app.Config.DB.DSN()))
	}

	if err != nil {
		panic(err)
	}
}
