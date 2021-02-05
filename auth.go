package main

import (
	"database/sql"
	"log"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type user struct {
	ID       int
	Username string
	Password string
}

func authRequired(c *gin.Context) {
	if sessions.Default(c).Get("id") == nil {
		c.AbortWithStatus(401)
	}
}

func getUser(c *gin.Context) (id int, username string, err error) {
	session := sessions.Default(c)
	sid := session.Get("id")
	username, _ = session.Get("username").(string)
	if universal {
		err = db.QueryRow("SELECT id FROM user WHERE uid = ?", sid).Scan(&id)
		return
	}
	id, _ = sid.(int)
	return
}

func login(c *gin.Context) {
	var login struct {
		Username, Password string
		Rememberme         bool
	}
	if err := c.BindJSON(&login); err != nil {
		c.String(400, "")
		return
	}
	login.Username = strings.ToLower(login.Username)

	var user user
	var message string
	if err := db.QueryRow(
		"SELECT id, username, password FROM user WHERE username = ?",
		login.Username,
	).Scan(&user.ID, &user.Username, &user.Password); err != nil {
		if strings.Contains(err.Error(), "doesn't exist") {
			restore("")
			c.String(503, "Detected first time running. Initialized the database.")
			return
		} else if err == sql.ErrNoRows {
			message = "Incorrect username"
		} else {
			log.Print(err)
			c.String(500, "Critical Error! Please contact your system administrator.")
			return
		}
	} else {
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password)); err != nil {
			if (err == bcrypt.ErrHashTooShort && user.Password != login.Password) ||
				err == bcrypt.ErrMismatchedHashAndPassword {
				message = "Incorrect password"
			} else if user.Password != login.Password {
				log.Print(err)
				c.String(500, "Critical Error! Please contact your system administrator.")
				return
			}
		}
		if message == "" {
			session := sessions.Default(c)
			session.Clear()
			session.Set("id", user.ID)
			session.Set("username", user.Username)

			if login.Rememberme {
				session.Options(sessions.Options{HttpOnly: true, MaxAge: 856400 * 365})
			} else {
				session.Options(sessions.Options{HttpOnly: true})
			}

			if err := session.Save(); err != nil {
				log.Print(err)
				c.String(500, "Internal Server Error")
				return
			}

			c.JSON(200, gin.H{"status": 1})
			return
		}
	}

	c.JSON(200, gin.H{"status": 0, "message": message})
}

func chgpwd(c *gin.Context) {
	var data struct{ Password, Password1, Password2 string }
	if err := c.BindJSON(&data); err != nil {
		c.String(400, "")
		return
	}

	session := sessions.Default(c)
	userID := session.Get("id")

	var oldPassword string
	if err := db.QueryRow("SELECT password FROM user WHERE id = ?", userID).Scan(&oldPassword); err != nil {
		log.Print(err)
		c.String(500, "")
		return
	}

	var message string
	var errorCode int
	err := bcrypt.CompareHashAndPassword([]byte(oldPassword), []byte(data.Password))
	switch {
	case err != nil && data.Password != oldPassword:
		if err == bcrypt.ErrHashTooShort || err == bcrypt.ErrMismatchedHashAndPassword {
			message = "Incorrect password."
			errorCode = 1
		} else {
			log.Print(err)
			c.String(500, "")
			return
		}
	case data.Password1 != data.Password2:
		message = "Confirm password doesn't match new password."
		errorCode = 2
	case data.Password1 == data.Password:
		message = "New password cannot be the same as your current password."
		errorCode = 2
	case data.Password1 == "":
		message = "New password cannot be blank."
	}

	if message == "" {
		newPassword, err := bcrypt.GenerateFromPassword([]byte(data.Password1), bcrypt.MinCost)
		if err != nil {
			log.Print(err)
			c.String(500, "")
			return
		}
		if _, err = db.Exec("UPDATE user SET password = ? WHERE id = ?", string(newPassword), userID); err != nil {
			log.Print(err)
			c.String(500, "")
			return
		}

		session.Clear()
		session.Options(sessions.Options{MaxAge: -1})
		if err := session.Save(); err != nil {
			log.Print(err)
			c.String(500, "")
			return
		}

		c.JSON(200, gin.H{"status": 1})
		return
	}

	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}
