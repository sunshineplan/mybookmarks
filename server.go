package main

import (
	"bytes"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
)

func run() {
	if *logPath != "" {
		f, err := os.OpenFile(*logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			log.Fatalln("Failed to open log file:", err)
		}
		gin.DefaultWriter = f
		gin.DefaultErrorWriter = f
		log.SetOutput(f)
	}

	if err := initDB(); err != nil {
		log.Fatalln("Failed to initialize database:", err)
	}

	router := gin.Default()
	server.Handler = router

	js, err := os.ReadFile(joinPath(dir(self), "public/build/bundle.js"))
	if err != nil {
		log.Fatal(err)
	}

	if *universal {
		var redisStore struct{ Endpoint, Password, Secret, API string }
		if err := meta.Get("account_redis", &redisStore); err != nil {
			log.Fatal(err)
		}

		js = bytes.ReplaceAll(js, []byte("@universal@"), []byte(redisStore.API))

		store, err := redis.NewStore(10, "tcp", redisStore.Endpoint, redisStore.Password, []byte(redisStore.Secret))
		if err != nil {
			log.Fatal(err)
		}
		if err := redis.SetKeyPrefix(store, "account_"); err != nil {
			log.Fatal(err)
		}
		router.Use(sessions.Sessions("universal", store))
	} else {
		js = bytes.ReplaceAll(js, []byte("@universal@"), nil)

		secret := make([]byte, 16)
		if _, err := rand.Read(secret); err != nil {
			log.Fatalln("Failed to get secret:", err)
		}
		router.Use(sessions.Sessions("session", cookie.NewStore(secret)))
	}

	if priv != nil {
		pubkey_bytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
		if err != nil {
			log.Fatal(err)
		}
		js = bytes.ReplaceAll(
			js, []byte("@pubkey@"),
			bytes.ReplaceAll(
				pem.EncodeToMemory(&pem.Block{
					Type:  "RSA PUBLIC KEY",
					Bytes: pubkey_bytes,
				}),
				[]byte{'\n'},
				nil,
			),
		)
	} else {
		js = bytes.ReplaceAll(js, []byte("@pubkey@"), nil)
	}

	if err := os.WriteFile(joinPath(dir(self), "public/build/script.js"), js, 0644); err != nil {
		log.Fatal(err)
	}

	router.StaticFS("/build", http.Dir(joinPath(dir(self), "public/build")))
	router.StaticFile("favicon.ico", joinPath(dir(self), "public/favicon.ico"))
	router.LoadHTMLFiles(joinPath(dir(self), "public/index.html"))
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})
	router.GET("/info", func(c *gin.Context) {
		id, username, _ := getUser(c)
		if username == "" {
			c.JSON(200, gin.H{})
			return
		}

		ch := make(chan error, 1)
		var categories []category
		go func() { var err error; categories, err = getCategory(id); ch <- err }()
		bookmarks, total, err := getBookmark(id)
		if err != nil {
			log.Println("Failed to get bookmarks:", err)
		}
		if err = <-ch; err != nil {
			log.Println("Failed to get categories:", err)
		}

		c.JSON(200, gin.H{"username": username, "categories": categories, "bookmarks": bookmarks, "total": total})
	})

	auth := router.Group("/")
	auth.POST("/login", login)
	auth.POST("/logout", authRequired, func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		session.Options(sessions.Options{MaxAge: -1})
		if err := session.Save(); err != nil {
			log.Print(err)
			c.String(500, "")
			return
		}
		c.String(200, "bye")
	})
	auth.POST("/chgpwd", authRequired, chgpwd)

	base := router.Group("/")
	base.Use(authRequired)
	base.POST("/bookmark/get", moreBookmark)
	base.POST("/bookmark/add", addBookmark)
	base.POST("/bookmark/edit/:id", editBookmark)
	base.POST("/bookmark/delete/:id", deleteBookmark)
	base.POST("/category/edit", editCategory)
	base.POST("/category/delete", deleteCategory)
	base.POST("/reorder", reorder)

	router.NoRoute(func(c *gin.Context) {
		c.Redirect(302, "/")
	})

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
