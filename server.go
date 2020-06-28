package main

import (
	"crypto/rand"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func loadTemplates() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	r.AddFromFiles("base.html", joinPath(dir(self), "templates/base.html"), joinPath(dir(self), "templates/root.html"))
	r.AddFromFiles("login.html", joinPath(dir(self), "templates/base.html"), joinPath(dir(self), "templates/auth/login.html"))
	r.AddFromFiles("setting.html", joinPath(dir(self), "templates/auth/setting.html"))

	includes, err := filepath.Glob(joinPath(dir(self), "templates/bookmark/*"))
	if err != nil {
		log.Fatal(err)
	}

	for _, include := range includes {
		r.AddFromFiles(filepath.Base(include), include)
	}
	return r
}

func run() {
	f, err := os.OpenFile(*logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		log.Fatal(err)
	}
	gin.DefaultWriter = io.MultiWriter(f)

	secret := make([]byte, 16)
	_, err = rand.Read(secret)
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.Use(sessions.Sessions("session", sessions.NewCookieStore(secret)))
	router.StaticFS("/static", http.Dir(joinPath(dir(self), "static")))
	router.HTMLRender = loadTemplates()
	router.GET("/", func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get("username")
		if username == nil {
			c.Redirect(302, "/auth/login")
			return
		}
		c.HTML(200, "base.html", gin.H{"user": username})
	})

	auth := router.Group("/auth")
	auth.GET("/login", func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user_id")
		if user != nil {
			c.Redirect(302, "/")
			return
		}
		c.HTML(200, "login.html", gin.H{"error": ""})
	})
	auth.POST("/login", login)
	auth.GET("/logout", authRequired, func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		session.Save()
		c.Redirect(302, "/auth/login")
	})
	auth.GET("/setting", authRequired, func(c *gin.Context) {
		c.HTML(200, "setting.html", nil)
	})
	auth.POST("/setting", authRequired, setting)

	base := router.Group("/")
	base.Use(authRequired)
	base.GET("/bookmark", getBookmark)
	base.GET("/bookmark/add", addBookmark)
	base.POST("/bookmark/add", doAddBookmark)
	base.GET("/bookmark/edit/:id", editBookmark)
	base.POST("/bookmark/edit/:id", doEditBookmark)
	base.POST("/bookmark/delete/:id", doDeleteBookmark)
	base.GET("/category/get", getCategory)
	base.GET("/category/add", func(c *gin.Context) {
		c.HTML(200, "category.html", gin.H{"id": 0})
	})
	base.POST("/category/add", doAddCategory)
	base.GET("/category/edit/:id", editCategory)
	base.POST("/category/edit/:id", doEditCategory)
	base.POST("/category/delete/:id", doDeleteCategory)
	base.POST("/reorder", reorder)

	if *unix != "" && runtime.GOOS == "linux" {
		if _, err := os.Stat(*unix); err == nil {
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
