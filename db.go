package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sunshineplan/metadata"
)

var dbConfig mysqlConfig
var dsn string

type mysqlConfig struct {
	Server   string
	Port     int
	Database string
	Username string
	Password string
}

func getDB() {
	m, err := metadata.Get("mybookmarks_mysql", &metadataConfig)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(m, &dbConfig)
	if err != nil {
		log.Fatalln(err)
	}
	dsn = fmt.Sprintf("%s:%s@%s:%d/%s", dbConfig.Username, dbConfig.Password, dbConfig.Server, dbConfig.Port, dbConfig.Database)
	dsn = "root:159357qw@/test" // test
}

func restore(filePath string, db *sql.DB) error {
	if filePath == "" {
		filePath = filepath.Join(filepath.Dir(self), "schema.sql")
	}
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	if db == nil {
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer db.Close()
	dropAll, err := ioutil.ReadFile(filepath.Join(filepath.Dir(self), "drop_all.sql"))
	if err != nil {
		log.Fatal(err)
	}
	tx, _ := db.Begin()
	_, err = tx.Exec(string(dropAll))
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Fatal(rollbackErr)
		}
		log.Fatal(err)
	}
	_, err = tx.Exec(string(file))
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Fatal(rollbackErr)
		}
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
	return nil
}

// Dump database
func Dump() string {
	tmpfile, err := ioutil.TempFile("", "tmp")
	if err != nil {
		log.Fatal(err)
	}
	args := []string{}
	args = append(args, fmt.Sprintf("-h%s", dbConfig.Server))
	args = append(args, fmt.Sprintf("-P%d", dbConfig.Port))
	args = append(args, fmt.Sprintf("-B%s", dbConfig.Database))
	args = append(args, fmt.Sprintf("-u%s", dbConfig.Username))
	args = append(args, fmt.Sprintf("-p%s", dbConfig.Password))
	args = append(args, "-C")
	args = append(args, fmt.Sprintf("-r%s", tmpfile.Name()))
	cmd := exec.Command("mysqldump", args...)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	return tmpfile.Name()
}
