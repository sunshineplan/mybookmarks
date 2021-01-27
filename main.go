package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sunshineplan/service"
	"github.com/sunshineplan/utils"
	"github.com/sunshineplan/utils/httpsvr"
	"github.com/sunshineplan/utils/metadata"
	"github.com/vharitonsky/iniflags"
)

var self string
var logPath string
var server httpsvr.Server
var meta metadata.Server

var svc = service.Service{
	Name: "MyBookmarks",
	Desc: "Instance to serve My Bookmarks",
	Exec: run,
	Options: service.Options{
		Dependencies: []string{"Wants=network-online.target", "After=network.target"},
		Others:       []string{"Environment=GIN_MODE=release"},
	},
}

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
	flag.StringVar(&svc.Options.UpdateURL, "update", "", "Update URL")
	exclude := flag.String("exclude", "", "Exclude Files")
	//flag.StringVar(&logPath, "log", joinPath(dir(self), "access.log"), "Log Path")
	flag.StringVar(&logPath, "log", "", "Log Path")
	iniflags.SetConfigFile(joinPath(dir(self), "config.ini"))
	iniflags.SetAllowMissingConfigFile(true)
	iniflags.Parse()

	svc.Options.ExcludeFiles = strings.Split(*exclude, ",")

	if err := initDB(); err != nil {
		log.Fatalln("Failed to load database config:", err)
	}

	if service.IsWindowsService() {
		svc.Run(false)
		return
	}

	switch flag.NArg() {
	case 0:
		run()
	case 1:
		switch flag.Arg(0) {
		case "run", "debug":
			run()
		case "install":
			err = svc.Install()
		case "remove":
			err = svc.Remove()
		case "start":
			err = svc.Start()
		case "stop":
			err = svc.Stop()
		case "update":
			err = svc.Update()
		case "backup":
			backup()
		case "init":
			if utils.Confirm("Do you want to initialize database?", 3) {
				restore("")
			}
		default:
			log.Fatalln("Unknown argument:", flag.Arg(0))
		}
	case 2:
		switch flag.Arg(0) {
		case "add":
			addUser(flag.Arg(1))
		case "delete":
			if utils.Confirm("Do you want to initialize database?", 3) {
				deleteUser(flag.Arg(1))
			}
		case "restore":
			if utils.Confirm("Do you want to initialize database?", 3) {
				restore(flag.Arg(1))
			}
		default:
			log.Fatalln("Unknown arguments:", strings.Join(flag.Args(), " "))
		}
	default:
		log.Fatalln("Unknown arguments:", strings.Join(flag.Args(), " "))
	}
	if err != nil {
		log.Fatalf("Failed to %s: %v", flag.Arg(0), err)
	}
}
