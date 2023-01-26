package main

import (
	"fmt"

	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/driver/database"
)

func main() {
	db, err := database.NewSqlite(config.Database{})
	if err != nil {
		panic(err)
	}

	rows, err := db.Query("SELECT 1")
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var data int
		rows.Scan(&data)
		fmt.Println(data)
	}

}
