package main

import (
	"database/sql"
	"errors"
	"log"
	"strings"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type user struct {
	id       int
	username string
	password string
}

// AuthRequired is a middleware to check the session
func AuthRequired(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		c.Redirect(302, "/auth/login")
		return
	}
	c.Next()
}

// Login is a handler that parses a form and checks for specific data
func Login(c *gin.Context) {
	session := sessions.Default(c)
	username := strings.TrimSpace(strings.ToLower(c.PostForm("username")))
	password := c.PostForm("password")

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()
	user := new(user)
	err = db.QueryRow("SELECT * FROM user WHERE username = ?", username).Scan(&user)
	//if err != nil {
	//	err = initDB(db)
	//	if err == nil {
	//		c.HTML(200, "/auth/login.html", gin.H{"error": "Detected first time running. Initialized the database."})
	//		return
	//	}
	//	c.HTML(200, "/auth/login.html", gin.H{"error": "Critical Error! Please contact your system administrator."})
	//	return
	//}
	if err != nil {
		err = errors.New("Incorrect username")
	} else if err = bcrypt.CompareHashAndPassword([]byte(user.password), []byte(password)); err != nil || user.password != password {
		err = errors.New("Incorrect password")
	}

	if err != nil {
		session.Clear()
		session.Set("user_id", user.id)
		if err := session.Save(); err != nil {
			c.JSON(500, gin.H{"error": "Failed to save session"})
			return
		}
		c.Redirect(302, "/")
		return
	}
	c.HTML(200, "/auth/login.html", gin.H{"error": err.Error()})
}

// Logout handler
func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	c.Redirect(302, "/")
}

// Setting is a handler that change user password
func Setting(c *gin.Context) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println(err)
		c.String(500, "")
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")

	password := c.PostForm("password")
	password1 := c.PostForm("password1")
	password2 := c.PostForm("password2")

	var Password string
	db.QueryRow("SELECT password FROM user WHERE id = ?", userID).Scan(&Password)

	var message string
	var errorCode int
	err = bcrypt.CompareHashAndPassword([]byte(Password), []byte(password))
	switch {
	case err != nil && Password != password:
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

	if message != "" {
		newPassword, err := bcrypt.GenerateFromPassword([]byte(password1), bcrypt.MinCost)
		if err != nil {
			log.Println(err)
			c.String(500, "")
			return
		}
		_, err = db.Exec("UPDATE user SET password = ? WHERE id = ?",
			string(newPassword), userID)
		if err != nil {
			log.Println(err)
			c.String(500, "")
			return
		}
		session.Clear()
		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}
