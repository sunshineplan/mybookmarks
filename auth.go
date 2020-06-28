package main

import (
	"database/sql"
	"log"
	"strings"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type user struct {
	ID       int
	Username string
	Password string
}

func authRequired(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		c.AbortWithStatus(401)
	}
}

func login(c *gin.Context) {
	session := sessions.Default(c)
	username := strings.TrimSpace(strings.ToLower(c.PostForm("username")))
	password := c.PostForm("password")

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println(err)
		c.HTML(200, "login.html", gin.H{"error": "Failed to connect to database."})
		return
	}
	defer db.Close()
	user := new(user)
	err = db.QueryRow("SELECT * FROM user WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.Password)
	var message string
	if err != nil {
		if strings.Contains(err.Error(), "doesn't exist") {
			err = restore("")
			if err == nil {
				c.HTML(200, "login.html", gin.H{"error": "Detected first time running. Initialized the database."})
				return
			}
			log.Println(err)
			c.HTML(200, "login.html", gin.H{"error": "Critical Error! Please contact your system administrator."})
			return
		}
		log.Println(err)
		message = "Incorrect username"
	} else if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil && user.Password != password {
		message = "Incorrect password"
	} else {
		session.Clear()
		session.Set("user_id", user.ID)
		session.Set("username", user.Username)

		rememberme := c.PostForm("rememberme")
		if rememberme == "on" {
			session.Options(sessions.Options{Path: "/", HttpOnly: true, MaxAge: 856400 * 365})
		} else {
			session.Options(sessions.Options{Path: "/", HttpOnly: true, MaxAge: 0})
		}

		if err := session.Save(); err != nil {
			log.Println(err)
			c.HTML(200, "login.html", gin.H{"error": "Failed to save session."})
			return
		}
		c.Redirect(302, "/")
		return
	}
	c.HTML(200, "login.html", gin.H{"error": message})
}

func setting(c *gin.Context) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println(err)
		c.String(503, "")
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")

	password := c.PostForm("password")
	password1 := c.PostForm("password1")
	password2 := c.PostForm("password2")

	var oldPassword string
	db.QueryRow("SELECT password FROM user WHERE id = ?", userID).Scan(&oldPassword)

	var message string
	var errorCode int
	err = bcrypt.CompareHashAndPassword([]byte(oldPassword), []byte(password))
	switch {
	case err != nil && password != oldPassword:
		log.Println(err)
		message = "Incorrect password."
		errorCode = 1
	case password1 != password2:
		message = "Confirm password doesn't match new password."
		errorCode = 2
	case password1 == password:
		message = "New password cannot be the same as your current password."
		errorCode = 2
	case password1 == "":
		message = "New password cannot be blank."
	}

	if message == "" {
		newPassword, err := bcrypt.GenerateFromPassword([]byte(password1), bcrypt.MinCost)
		if err != nil {
			log.Println(err)
			c.String(500, "")
			return
		}
		_, err = db.Exec("UPDATE user SET password = ? WHERE id = ?", string(newPassword), userID)
		if err != nil {
			log.Println(err)
			c.String(500, "")
			return
		}
		session.Clear()
		if err := session.Save(); err != nil {
			log.Println(err)
			c.String(500, "")
			return
		}
		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}
