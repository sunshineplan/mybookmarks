package main

import (
	"database/sql"
	"io/ioutil"
	"path/filepath"
)

type mysql struct {
	server   string
	port     int
	database string
	username string
	password string
}

func initDB(db *sql.DB) error {
	file, err := ioutil.ReadFile(filepath.Join(filepath.Dir(self), "drop_all.sql"))
	if err != nil {
		return err
	}
	_, err = db.Exec(string(file))
	if err != nil {
		return err
	}
	file, err = ioutil.ReadFile(filepath.Join(filepath.Dir(self), "schema.sql"))
	if err != nil {
		return err
	}
	_, err = db.Exec(string(file))
	if err != nil {
		return err
	}
	return nil
}
