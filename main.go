package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sunshineplan/metadata"
	"github.com/vharitonsky/iniflags"
)

// OS is the running program's operating system
const OS = runtime.GOOS

var metadataConfig metadata.Config

var self string
var unix, host, port, logPath *string

var (
	joinPath = filepath.Join
	dir      = filepath.Dir
)

func main() {
	var err error
	self, err = os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&metadataConfig.Server, "server", "", "Metadata Server Address")
	flag.StringVar(&metadataConfig.VerifyHeader, "header", "", "Verify Header Header Name")
	flag.StringVar(&metadataConfig.VerifyValue, "value", "", "Verify Header Value")
	unix = flag.String("unix", "", "UNIX-domain Socket")
	host = flag.String("host", "127.0.0.1", "Server Host")
	port = flag.String("port", "12345", "Server Port")
	logPath = flag.String("log", joinPath(dir(self), "access.log"), "Log Path")
	iniflags.SetConfigFile(joinPath(dir(self), "config.ini"))
	iniflags.SetAllowMissingConfigFile(true)
	iniflags.Parse()
	getDB()

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
			log.Println("Start!")
			err := restore("")
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Done!")
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
			log.Println("Start!")
			err := restore(flag.Arg(1))
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Done!")
		default:
			log.Fatalf("Unknown arguments: %s", strings.Join(flag.Args(), " "))
		}
	default:
		log.Fatalf("Unknown arguments: %s", strings.Join(flag.Args(), " "))
	}
}
