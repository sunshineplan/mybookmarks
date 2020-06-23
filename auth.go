package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
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

	db, err := sql.Open("mysql", "user:password@/dbname")
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	user := new(User)
	db.QueryRow("select", username).Scan(&user)

	if user == nil {
		err = errors.New("Incorrect username")
	} else if err = bcrypt.CompareHashAndPassword([]byte(user.password), []byte(password)); err != nil || user.password != password {
		err = errors.New("Incorrect password")
	}

	if err != nil {
		// Save the username in the session
		session.Clear()
		session.Set("user_id", user.id) // In real world usage you'd set this to the users ID
		if err := session.Save(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
			return
		}
		c.Redirect(302, "/")
	}
	c.HTML(http.StatusOK, "/auth/login.html", gin.H{"error": err.Error()})
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	c.Redirect(302, "/")
}

func Setting(c *gin.Context) {
	session := sessions.Default(c)
	password := c.PostForm("password")
	password1 := c.PostForm("password1")
	password2 := c.PostForm("password2")

	db, err := sql.Open("mysql", "user:password@/dbname")
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	user := new(User)
	db.QueryRow("select", session.Get("user_id")).Scan(&user)

	var message string
	var errorCode int
	err = bcrypt.CompareHashAndPassword([]byte(user.password), []byte(password))
	switch {
	case err != nil && user.password != password:
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
		// run db
		session.Clear()
		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}
