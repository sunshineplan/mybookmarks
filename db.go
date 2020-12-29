package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/sunshineplan/utils/database/mysql"
)

var dbConfig mysql.Config
var db *sql.DB

func initDB() error {
	if err := meta.Get("mybookmarks_mysql", &dbConfig); err != nil {
		return err
	}
	return nil
}

func getDB() {
	var err error
	db, err = dbConfig.Open()
	if err != nil {
		log.Fatalln("Failed to connect to database:", err)
	}
	db.SetConnMaxLifetime(time.Minute * 1)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
}
