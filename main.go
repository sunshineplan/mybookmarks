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

var metadataConfig metadata.Config

var self string
var unix, host, port, logPath *string

func main() {
	self, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&metadataConfig.Server, "server", "", "Metadata Server Address")
	flag.StringVar(&metadataConfig.VerifyHeader, "header", "", "Verify Header Header Name")
	flag.StringVar(&metadataConfig.VerifyValue, "value", "", "Verify Header Value")
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
		case "init":
			err := restore("", nil)
			if err == nil {
				log.Println("Done.")
			}
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
			err := restore(flag.Arg(1), nil)
			if err == nil {
				log.Println("Done.")
			}
		default:
			log.Fatalf("Unknown arguments: %s", strings.Join(flag.Args(), " "))
		}
	default:
		log.Fatalf("Unknown arguments: %s", strings.Join(flag.Args(), " "))
	}
}
