package main

import (
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/metadata"
	"github.com/vharitonsky/iniflags"
)

var config metadata.Config
var mariadb mysql
var self string

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
	unix := flag.String("unix", "", "Server Host")
	host := flag.String("host", "127.0.0.1", "Server Host")
	port := flag.String("port", "12345", "Server Port")
	logPath := flag.String("log", "/var/log/app/mybookmarks-go.log", "Log Path")
	iniflags.SetConfigFile(filepath.Join(filepath.Dir(self), "config.ini"))
	iniflags.SetAllowMissingConfigFile(true)
	iniflags.Parse()

	f, _ := os.OpenFile(*logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	gin.DefaultWriter = io.MultiWriter(f)

	router := gin.Default()
	router.Use(sessions.Sessions("mysession", sessions.NewCookieStore([]byte("secret"))))
	router.StaticFS("/static", http.Dir(filepath.Join(filepath.Dir(self), "static")))
	router.LoadHTMLGlob(filepath.Join(filepath.Dir(self), "templates/**/*"))

	auth := router.Group("/auth")
	auth.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "auth/login.html", nil)
	})
	auth.POST("/login", Login)
	auth.POST("/logout", AuthRequired, Logout)
	auth.GET("/setting", AuthRequired, func(c *gin.Context) {
		c.HTML(http.StatusOK, "auth/setting.html", nil)
	})
	auth.POST("/setting", AuthRequired, Setting)

	base := router.Group("/")
	base.Use(AuthRequired)
	base.GET("/", func(c *gin.Context) { c.HTML(http.StatusOK, "base.html", nil) })
	base.GET("/bookmark", bookmark)
	base.GET("/:mode/:action", modeAction)
	base.POST("/:mode/:action", doModeAction)
	base.GET("/:mode/:action/:id", modeActionID)
	base.POST("/:mode/:action/:id", doModeActionID)
	base.POST("/reorder", reorder)

	if *unix != "" && runtime.GOOS == "linux" {
		if _, err = os.Stat(*unix); err == nil {
			err = os.Remove(*unix)
			if err != nil {
				log.Fatal(err)
			}
		}

		listener, err := net.Listen("unix", *unix)
		if err != nil {
			log.Fatal(err)
		}

		idleConnsClosed := make(chan struct{})
		go func() {
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			<-quit

			if err := listener.Close(); err != nil {
				log.Printf("HTTP Listener close: %v", err)
			}
			if err := os.Remove(*unix); err != nil {
				log.Printf("Remove socket file: %v", err)
			}
			close(idleConnsClosed)
		}()

		if err = os.Chmod(*unix, 0666); err != nil {
			log.Fatal(err)
		}

		http.Serve(listener, router)
		<-idleConnsClosed
	} else {
		router.Run(*host + ":" + *port)
	}
}
