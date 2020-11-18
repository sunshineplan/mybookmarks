package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"runtime"
	"strings"
)

var dbc mysqlConfig
var dsn string

type mysqlConfig struct {
	Server   string
	Port     int
	Database string
	Username string
	Password string
}

func initDB() error {
	if err := meta.Get("mybookmarks_mysql", &dbc); err != nil {
		return err
	}
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbc.Username, dbc.Password, dbc.Server, dbc.Port, dbc.Database)
	return nil
}

func getDB() (*sql.DB, error) {
	return sql.Open("mysql", dsn)
}

func execScript(file string) {
	var cmd, arg string
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		arg = "/c"
	case "linux":
		cmd = "bash"
		arg = "-c"
	default:
		log.Fatal("Unsupported operating system.")
	}

	var args []string
	args = append(args, "mysql")
	args = append(args, fmt.Sprintf("%s", dbc.Database))
	args = append(args, fmt.Sprintf("-h%s", dbc.Server))
	args = append(args, fmt.Sprintf("-P%d", dbc.Port))
	args = append(args, fmt.Sprintf("-u%s", dbc.Username))
	args = append(args, fmt.Sprintf("-p%s", dbc.Password))
	args = append(args, "<")
	args = append(args, file)

	c := exec.Command(cmd, arg, strings.Join(args, " "))
	if err := c.Run(); err != nil {
		log.Fatalln("Failed to execute mysql script:", err)
	}
}

func dump() string {
	tmpfile, err := ioutil.TempFile("", "tmp")
	if err != nil {
		log.Fatalln("Failed to create temporary file:", err)
	}
	tmpfile.Close()

	var cmd, arg string
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		arg = "/c"
	case "linux":
		cmd = "bash"
		arg = "-c"
	default:
		log.Fatal("Unsupported operating system.")
	}

	var args []string
	args = append(args, "mysqldump")
	args = append(args, fmt.Sprintf("-h%s", dbc.Server))
	args = append(args, fmt.Sprintf("-P%d", dbc.Port))
	args = append(args, fmt.Sprintf("-u%s", dbc.Username))
	args = append(args, fmt.Sprintf("-p%s", dbc.Password))
	args = append(args, fmt.Sprintf("-r%s", tmpfile.Name()))
	args = append(args, "--add-drop-database")
	args = append(args, "--add-drop-trigger")
	args = append(args, "-CB")
	args = append(args, fmt.Sprintf("%s", dbc.Database))

	dump := exec.Command(cmd, arg, strings.Join(args, " "))
	if err := dump.Run(); err != nil {
		log.Fatalln("Failed to run backup command:", err)
	}
	return tmpfile.Name()
}
