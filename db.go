package main

import (
	"database/sql"
	"time"

	"github.com/sunshineplan/utils/database/mysql"
)

var dbConfig mysql.Config
var db *sql.DB

func initDB() (err error) {
	if err = meta.Get("mybookmarks_mysql", &dbConfig); err != nil {
		return
	}

	db, err = dbConfig.Open()
	if err != nil {
		return
	}
	db.SetConnMaxLifetime(time.Minute * 1)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return
}
