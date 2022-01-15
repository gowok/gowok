package gowok

import (
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func (app *App) dbConnect() {
	var err error
	if app.Config.DB.Driver == DriverMySQL {
		app.db, err = gorm.Open(mysql.Open(app.Config.DB.DSN()))
	} else if app.Config.DB.Driver == DriverPostgreSQL {
		app.db, err = gorm.Open(postgres.Open(app.Config.DB.DSN()))
	}

	if err != nil {
		panic(err)
	}
}
