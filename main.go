package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sunshineplan/metadata"
	"github.com/vharitonsky/iniflags"
)

var config metadata.Config
var mariadb mysql
var self string
var unix, host, port, logPath *string

func init() {
	c, err := metadata.Get("mybookmarks_mysql", &config)
	if err != nil {
		log.Fatal(err)
	}
	mariadb = c.(mysql)
}

func main() {
	self, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&config.Server, "server", "", "Metadata Server Address")
	flag.StringVar(&config.VerifyHeader, "header", "", "Verify Header Header Name")
	flag.StringVar(&config.VerifyValue, "value", "", "Verify Header Value")
	unix = flag.String("unix", "", "Server Host")
	host = flag.String("host", "127.0.0.1", "Server Host")
	port = flag.String("port", "12345", "Server Port")
	logPath = flag.String("log", "/var/log/app/mybookmarks-go.log", "Log Path")
	iniflags.SetConfigFile(filepath.Join(filepath.Dir(self), "config.ini"))
	iniflags.SetAllowMissingConfigFile(true)
	iniflags.Parse()

	switch flag.NArg() {
	case 0:
		run()
	case 1:
		switch flag.Arg(0) {
		case "run":
			run()
		case "backup":
			backup()
		default:
			log.Fatalf("Unknown argument: %s", flag.Arg(0))
		}
	case 2:
		switch flag.Arg(0) {
		case "add":
			addUser(flag.Arg(1))
		case "delete":
			deleteUser(flag.Arg(1))
		case "restore":
			restore(flag.Arg(1))
		default:
			log.Fatalf("Unknown arguments: %s", strings.Join(flag.Args(), " "))
		}
	default:
		log.Fatalf("Unknown arguments: %s", strings.Join(flag.Args(), " "))
	}
}
