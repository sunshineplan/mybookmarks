package main

import (
	"database/sql"
	"log"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/utils/password"
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
		ok, err := password.Compare(user.Password, login.Password, false)
		if err != nil {
			log.Print(err)
			c.String(500, "Internal Server Error")
			return
		} else if !ok {
			message = "Incorrect password"
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
	newPassword, err := password.Change(oldPassword, data.Password, data.Password1, data.Password2, false)
	if err != nil {
		message = err.Error()
		switch err {
		case password.ErrIncorrectPassword:
			errorCode = 1
		case password.ErrConfirmPasswordNotMatch, password.ErrSamePassword:
			errorCode = 2
		case password.ErrBlankPassword:
		default:
			log.Print(err)
			c.String(500, "Internal Server Error")
			return
		}
	}

	if message == "" {
		if _, err = db.Exec("UPDATE user SET password = ? WHERE id = ?", newPassword, userID); err != nil {
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
