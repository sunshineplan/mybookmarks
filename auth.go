package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/database/mongodb"
	"github.com/sunshineplan/password"
)

type user struct {
	ID       string `json:"_id"`
	Username string
	Password string
	Last     string
}

func getUser(session sessions.Session) (usr user, err error) {
	id := session.Get("id")
	if id == nil {
		return
	}

	var filter any
	if *universal {
		filter = mongodb.M{"uid": id}
	} else {
		id, _ := accountClient.ObjectID(id.(string))
		filter = mongodb.M{"_id": id.Interface()}
	}
	mu.Lock()
	defer mu.Unlock()
	err = accountClient.FindOne(filter, nil, &usr)
	return
}

func authRequired(c *gin.Context) {
	if user, err := getUser(sessions.Default(c)); user.ID == "" || err == mongodb.ErrNoDocuments {
		c.AbortWithStatus(401)
	} else if user.ID != "" {
		c.Set("id", user.ID)
		c.Set("last", user.Last)
	} else {
		c.AbortWithStatus(500)
	}
}

func login(c *gin.Context) {
	var login struct {
		Username, Password string
		Rememberme         bool
	}
	if err := c.BindJSON(&login); err != nil {
		svc.Print(err)
		return
	}
	login.Username = strings.ToLower(login.Username)

	if password.IsMaxAttempts(c.ClientIP() + login.Username) {
		c.JSON(200, gin.H{"status": 0, "message": fmt.Sprintf("Max retries exceeded (%d)", maxRetry)})
		return
	}

	var user user
	var message string
	if err := accountClient.FindOne(mongodb.M{"username": login.Username}, nil, &user); err != nil {
		if err == mongodb.ErrNoDocuments {
			message = "Incorrect username"
		} else {
			svc.Print(err)
			c.String(500, "Critical Error! Please contact your system administrator.")
			c.Abort()
			return
		}
	} else {
		if priv == nil {
			_, err = password.Compare(c.ClientIP()+login.Username, user.Password, login.Password, false)
		} else {
			_, err = password.CompareRSA(c.ClientIP()+login.Username, user.Password, login.Password, false, priv)
		}
		if err != nil {
			if errors.Is(err, password.ErrIncorrectPassword) {
				message = err.Error()
			} else {
				svc.Print(err)
				c.AbortWithStatus(500)
				return
			}
		}

		if message == "" {
			session := sessions.Default(c)
			session.Clear()
			session.Set("id", user.ID)

			if login.Rememberme {
				session.Options(sessions.Options{HttpOnly: true, MaxAge: 856400 * 365})
			} else {
				session.Options(sessions.Options{HttpOnly: true})
			}

			if err := session.Save(); err != nil {
				svc.Print(err)
				c.AbortWithStatus(500)
				return
			}

			c.JSON(200, gin.H{"status": 1})
			return
		}
	}

	c.JSON(200, gin.H{"status": 0, "message": message})
}

func chgpwd(c *gin.Context) {
	session := sessions.Default(c)
	userID, username := session.Get("id"), session.Get("username")
	if userID == nil || username == nil {
		c.AbortWithStatus(401)
		return
	}
	id, _ := accountClient.ObjectID(userID.(string))

	if password.IsMaxAttempts(c.ClientIP() + username.(string)) {
		c.JSON(200, gin.H{"status": 0, "message": fmt.Sprintf("Max retries exceeded (%d)", maxRetry), "error": 1})
		return
	}

	var data struct{ Password, Password1, Password2 string }
	if err := c.BindJSON(&data); err != nil {
		svc.Print(err)
		return
	}

	var user user
	if err := accountClient.FindOne(mongodb.M{"_id": id.Interface()}, nil, &user); err != nil {
		svc.Print(err)
		c.AbortWithStatus(500)
		return
	}

	var err error
	var message, newPassword string
	var errorCode int
	if priv == nil {
		newPassword, err = password.Change(
			c.ClientIP()+user.Username, user.Password, data.Password, data.Password1, data.Password2, false,
		)
	} else {
		newPassword, err = password.ChangeRSA(
			c.ClientIP()+user.Username, user.Password, data.Password, data.Password1, data.Password2, false, priv,
		)
	}
	if err != nil {
		message = err.Error()
		switch {
		case errors.Is(err, password.ErrIncorrectPassword):
			errorCode = 1
		case err == password.ErrConfirmPasswordNotMatch || err == password.ErrSamePassword:
			errorCode = 2
		case err == password.ErrBlankPassword:
		default:
			svc.Print(err)
			c.AbortWithStatus(500)
			return
		}
	}

	if message == "" {
		if _, err := accountClient.UpdateOne(
			mongodb.M{"_id": id.Interface()},
			mongodb.M{"$set": mongodb.M{"password": newPassword}},
			nil,
		); err != nil {
			svc.Print(err)
			c.AbortWithStatus(500)
			return
		}

		session.Clear()
		session.Options(sessions.Options{MaxAge: -1})
		if err := session.Save(); err != nil {
			svc.Print(err)
			c.AbortWithStatus(500)
			return
		}

		c.JSON(200, gin.H{"status": 1})
		return
	}

	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}
