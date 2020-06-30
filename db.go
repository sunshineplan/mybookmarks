package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"

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
		log.Fatalf("Failed to get mybookmarks_mysql metadata: %v", err)
	}
	err = json.Unmarshal(m, &dbConfig)
	if err != nil {
		log.Fatalf("Failed to unmarshal json: %v", err)
	}
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbConfig.Username, dbConfig.Password, dbConfig.Server, dbConfig.Port, dbConfig.Database)
}

func restore(filePath string) {
	log.Println("Start!")
	if filePath == "" {
		filePath = joinPath(dir(self), "scripts/schema.sql")
	}
	dropAll := joinPath(dir(self), "scripts/drop_all.sql")

	var cmd, arg string
	switch OS {
	case "windows":
		cmd = "cmd"
		arg = "/c"
	case "linux":
		cmd = "bash"
		arg = "-c"
	default:
		log.Fatal("Unsupported operating system.")
	}

	args := []string{}
	args = append(args, "mysql")
	args = append(args, fmt.Sprintf("%s", dbConfig.Database))
	args = append(args, fmt.Sprintf("-h%s", dbConfig.Server))
	args = append(args, fmt.Sprintf("-P%d", dbConfig.Port))
	args = append(args, fmt.Sprintf("-u%s", dbConfig.Username))
	args = append(args, fmt.Sprintf("-p%s", dbConfig.Password))
	args = append(args, "<")

	drop := exec.Command(cmd, arg, strings.Join(append(args, dropAll), " "))
	if err := drop.Run(); err != nil {
		log.Fatalf("Failed to run drop all script: %v", err)
		return
	}

	restore := exec.Command(cmd, arg, strings.Join(append(args, filePath), " "))
	if err := restore.Run(); err != nil {
		log.Fatalf("Failed to run restore script: %v", err)
	}
	log.Println("Done!")
}

// Dump database
func Dump() string {
	tmpfile, err := ioutil.TempFile("", "tmp")
	if err != nil {
		log.Fatalf("Failed to create temporary file: %v", err)
	}

	var cmd, arg string
	switch OS {
	case "windows":
		cmd = "cmd"
		arg = "/c"
	case "linux":
		cmd = "bash"
		arg = "-c"
	default:
		log.Fatal("Unsupported operating system.")
	}

	args := []string{}
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

	dump := exec.Command(cmd, arg, strings.Join(args, " "))
	if err := dump.Run(); err != nil {
		log.Fatalf("Failed to run backup command: %v", err)
	}
	return tmpfile.Name()
}
