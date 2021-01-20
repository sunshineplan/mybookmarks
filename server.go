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
			log.Fatalln("Failed to open log file:", err)
		}
		gin.DefaultWriter = f
		gin.DefaultErrorWriter = f
		log.SetOutput(f)
	}

	secret := make([]byte, 16)
	if _, err := rand.Read(secret); err != nil {
		log.Fatalln("Failed to get secret:", err)
	}

	router := gin.Default()
	server.Handler = router
	router.Use(sessions.Sessions("session", sessions.NewCookieStore(secret)))
	router.StaticFS("/build", http.Dir(joinPath(dir(self), "public/build")))
	router.StaticFile("favicon.ico", joinPath(dir(self), "public/favicon.ico"))
	router.LoadHTMLFiles(joinPath(dir(self), "public/index.html"))
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})
	router.GET("/info", func(c *gin.Context) {
		userID := sessions.Default(c).Get("userID")
		username, _ := getUser(userID)
		categories, _ := getCategory(userID)
		bookmarks, _ := getBookmark(userID)
		c.JSON(200, gin.H{"username": username, "categories": categories, "bookmarks": bookmarks})
	})

	auth := router.Group("/")
	auth.POST("/login", login)
	auth.GET("/logout", authRequired, func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		session.Save()
		c.Redirect(302, "/")
	})
	auth.POST("/setting", authRequired, setting)

	base := router.Group("/")
	base.Use(authRequired)
	base.POST("/bookmark/get", moreBookmark)
	base.POST("/bookmark/add", addBookmark)
	base.POST("/bookmark/edit/:id", editBookmark)
	base.POST("/bookmark/delete/:id", deleteBookmark)
	base.POST("/category/add", addCategory)
	base.POST("/category/edit/:id", editCategory)
	base.POST("/category/delete/:id", deleteCategory)
	base.POST("/reorder", reorder)

	router.NoRoute(func(c *gin.Context) {
		c.Redirect(302, "/")
	})

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
