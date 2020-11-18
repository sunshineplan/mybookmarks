package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sunshineplan/utils/httpsvr"
	"github.com/sunshineplan/utils/metadata"
	"github.com/vharitonsky/iniflags"
)

var self string
var logPath string
var server httpsvr.Server
var meta metadata.Server

var (
	joinPath = filepath.Join
	dir      = filepath.Dir
)

func main() {
	var err error
	self, err = os.Executable()
	if err != nil {
		log.Fatalln("Failed to get self path:", err)
	}

	flag.StringVar(&meta.Addr, "server", "", "Metadata Server Address")
	flag.StringVar(&meta.Header, "header", "", "Verify Header Header Name")
	flag.StringVar(&meta.Value, "value", "", "Verify Header Value")
	flag.StringVar(&server.Unix, "unix", "", "UNIX-domain Socket")
	flag.StringVar(&server.Host, "host", "0.0.0.0", "Server Host")
	flag.StringVar(&server.Port, "port", "12345", "Server Port")
	//flag.StringVar(&logPath, "log", joinPath(dir(self), "access.log"), "Log Path")
	flag.StringVar(&logPath, "log", "", "Log Path")
	iniflags.SetConfigFile(joinPath(dir(self), "config.ini"))
	iniflags.SetAllowMissingConfigFile(true)
	iniflags.Parse()
	if err := initDB(); err != nil {
		log.Fatalln("Failed to load database config:", err)
	}

	switch flag.NArg() {
	case 0:
		run()
	case 1:
		switch flag.Arg(0) {
		case "run":
			run()
		case "backup":
			backup()
		case "init":
			restore("")
		default:
			log.Fatalln("Unknown argument:", flag.Arg(0))
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
			log.Fatalln("Unknown arguments:", strings.Join(flag.Args(), " "))
		}
	default:
		log.Fatalln("Unknown arguments:", strings.Join(flag.Args(), " "))
	}
}
