package main

import (
	"crypto/rand"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func run() {
	if logPath != "" {
		f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		gin.DefaultWriter = f
		gin.DefaultErrorWriter = f
		log.SetOutput(f)
	}

	secret := make([]byte, 16)
	if _, err := rand.Read(secret); err != nil {
		log.Fatalf("Failed to get secret: %v", err)
	}

	router := gin.Default()
	server.Handler = router
	router.Use(sessions.Sessions("session", sessions.NewCookieStore(secret)))
	router.StaticFS("/static", http.Dir(joinPath(dir(self), "static")))
	router.LoadHTMLFiles(joinPath(dir(self), "templates/index.html"))
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{"user": sessions.Default(c).Get("username")})
	})

	auth := router.Group("/")
	auth.POST("/login", login)
	auth.POST("/logout", authRequired, func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		session.Save()
		c.JSON(200, gin.H{"status": 1})
	})
	auth.POST("/setting", authRequired, setting)

	base := router.Group("/")
	base.Use(authRequired)
	base.POST("/bookmark/get", getBookmark)
	base.POST("/bookmark/add", addBookmark)
	base.POST("/bookmark/edit/:id", editBookmark)
	base.POST("/bookmark/delete/:id", deleteBookmark)
	base.POST("/category/get", getCategory)
	base.POST("/category/add", addCategory)
	base.POST("/category/edit/:id", editCategory)
	base.POST("/category/delete/:id", deleteCategory)
	base.POST("/reorder", reorder)

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
