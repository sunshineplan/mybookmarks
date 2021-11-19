package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sunshineplan/password"
	"github.com/sunshineplan/service"
	"github.com/sunshineplan/utils"
	"github.com/sunshineplan/utils/httpsvr"
	"github.com/sunshineplan/utils/metadata"
	"github.com/vharitonsky/iniflags"
)

var self string
var meta metadata.Server
var priv *rsa.PrivateKey

var server = httpsvr.New()
var svc = service.Service{
	Name:     "MyBookmarks",
	Desc:     "Instance to serve My Bookmarks",
	Exec:     run,
	TestExec: test,
	Options: service.Options{
		Dependencies: []string{"Wants=network-online.target", "After=network.target"},
		Environment:  map[string]string{"GIN_MODE": "release"},
	},
}

var (
	joinPath = filepath.Join
	dir      = filepath.Dir
)

func init() {
	var err error
	self, err = os.Executable()
	if err != nil {
		log.Fatalln("Failed to get self path:", err)
	}
}

var (
	maxRetry  = flag.Int("retry", 5, "Max number of retries on wrong password")
	universal = flag.Bool("universal", false, "Use Universal account id or not")
	pemPath   = flag.String("pem", "", "PEM File Path")
	exclude   = flag.String("exclude", "", "Exclude Files")
	logPath   = flag.String("log", "", "Log Path")
	// logPath = flag.String("log", joinPath(dir(self), "access.log"), "Log Path")
)

func main() {
	flag.StringVar(&meta.Addr, "server", "", "Metadata Server Address")
	flag.StringVar(&meta.Header, "header", "", "Verify Header Header Name")
	flag.StringVar(&meta.Value, "value", "", "Verify Header Value")
	flag.StringVar(&server.Unix, "unix", "", "UNIX-domain Socket")
	flag.StringVar(&server.Host, "host", "0.0.0.0", "Server Host")
	flag.StringVar(&server.Port, "port", "12345", "Server Port")
	flag.StringVar(&svc.Options.UpdateURL, "update", "", "Update URL")
	iniflags.SetConfigFile(joinPath(dir(self), "config.ini"))
	iniflags.SetAllowMissingConfigFile(true)
	iniflags.SetAllowUnknownFlags(true)
	iniflags.Parse()

	password.SetMaxAttempts(*maxRetry)
	if *pemPath != "" {
		b, err := os.ReadFile(*pemPath)
		if err != nil {
			log.Fatal(err)
		}
		block, _ := pem.Decode(b)
		if block == nil {
			log.Fatal("no PEM data is found")
		}
		priv, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			log.Fatal(err)
		}
	}
	svc.Options.ExcludeFiles = strings.Split(*exclude, ",")

	if service.IsWindowsService() {
		svc.Run(false)
		return
	}

	var err error
	switch flag.NArg() {
	case 0:
		run()
	case 1:
		switch flag.Arg(0) {
		case "run":
			svc.Run(false)
		case "debug":
			svc.Run(true)
		case "test":
			err = svc.Test()
		case "install":
			err = svc.Install()
		case "remove":
			err = svc.Remove()
		case "start":
			err = svc.Start()
		case "stop":
			err = svc.Stop()
		case "restart":
			err = svc.Restart()
		case "update":
			err = svc.Update()
		case "add", "delete", "backup", "restore":
			log.Fatalf("%s need two arguments", flag.Arg(0))
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
		case "backup":
			backup(flag.Arg(1))
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
