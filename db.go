package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"

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
		log.Fatal(err)
	}
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbConfig.Username, dbConfig.Password, dbConfig.Server, dbConfig.Port, dbConfig.Database)
}

func restore(filePath string) error {
	if filePath == "" {
		filePath = joinPath(dir(self), "schema.sql")
	}
	dropAll := joinPath(dir(self), "drop_all.sql")

	args := []string{}
	args = append(args, "/c")
	args = append(args, "mysql")
	args = append(args, fmt.Sprintf("%s", dbConfig.Database))
	args = append(args, fmt.Sprintf("-h%s", dbConfig.Server))
	args = append(args, fmt.Sprintf("-P%d", dbConfig.Port))
	args = append(args, fmt.Sprintf("-u%s", dbConfig.Username))
	args = append(args, fmt.Sprintf("-p%s", dbConfig.Password))
	args = append(args, "<")

	drop := exec.Command("cmd", append(args, dropAll)...)
	if err := drop.Run(); err != nil {
		log.Fatal(err)
		return err
	}

	restore := exec.Command("cmd", append(args, filePath)...)
	if err := restore.Run(); err != nil {
		log.Fatal(err)
		return err
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
	args = append(args, "/c")
	args = append(args, "mysqldump")
	args = append(args, fmt.Sprintf("-h%s", dbConfig.Server))
	args = append(args, fmt.Sprintf("-P%d", dbConfig.Port))
	args = append(args, fmt.Sprintf("-u%s", dbConfig.Username))
	args = append(args, fmt.Sprintf("-p%s", dbConfig.Password))
	args = append(args, fmt.Sprintf("-r%s", tmpfile.Name()))
	args = append(args, "--add-drop-database")
	args = append(args, "--add-drop-trigger")
	args = append(args, "-CB")
	args = append(args, fmt.Sprintf("%s", dbConfig.Database))
	cmd := exec.Command("cmd", args...)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	return tmpfile.Name()
}
