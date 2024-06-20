package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/metadata"
	"github.com/sunshineplan/password"
	"github.com/sunshineplan/service"
	"github.com/sunshineplan/utils"
	"github.com/sunshineplan/utils/flags"
	"github.com/sunshineplan/utils/httpsvr"
)

var (
	self string
	priv *rsa.PrivateKey

	server = httpsvr.New()
	svc    = service.New()
	meta   metadata.Server

	joinPath = filepath.Join
	dir      = filepath.Dir
)

func init() {
	var err error
	self, err = os.Executable()
	if err != nil {
		svc.Fatalln("Failed to get self path:", err)
	}
	svc.Name = "MyBookmarks"
	svc.Desc = "Instance to serve My Bookmarks"
	svc.Exec = run
	svc.TestExec = test
	svc.Options = service.Options{
		Dependencies:       []string{"Wants=network-online.target", "After=network.target"},
		Environment:        map[string]string{"GIN_MODE": "release"},
		RemoveBeforeUpdate: []string{"dist/assets"},
		ExcludeFiles:       []string{"scripts/mybookmarks.conf"},
	}
	svc.RegisterCommand("add", "add user", func(arg ...string) error {
		return addUser(arg[0])
	}, 1, true)
	svc.RegisterCommand("delete", "delete user", func(arg ...string) error {
		if utils.Confirm("Do you want to delete this user?", 3) {
			return deleteUser(arg[0])
		}
		return nil
	}, 1, true)
}

var (
	maxRetry  = flag.Int("retry", 5, "Max number of retries on wrong password")
	universal = flag.Bool("universal", false, "Use Universal account id or not")
	pemPath   = flag.String("pem", "", "PEM File Path")
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
	flag.StringVar(&svc.Options.PIDFile, "pid", "/var/run/mybookmarks.pid", "PID file path")
	flags.SetConfigFile(joinPath(dir(self), "config.ini"))
	flags.Parse()

	password.SetMaxAttempts(*maxRetry)
	if *pemPath != "" {
		b, err := os.ReadFile(*pemPath)
		if err != nil {
			svc.Fatal(err)
		}
		block, _ := pem.Decode(b)
		if block == nil {
			svc.Fatal("no PEM data is found")
		}
		priv, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			svc.Fatal(err)
		}
		password.SetKey(priv)
	}
	if *logPath != "" {
		svc.SetLogger(*logPath, "", log.LstdFlags)
		gin.DefaultWriter = svc.Logger
		gin.DefaultErrorWriter = svc.Logger
	}

	if err := svc.ParseAndRun(flag.Args()); err != nil {
		svc.Fatal(err)
	}
}
