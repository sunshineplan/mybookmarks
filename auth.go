package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/database/mongodb"
	"github.com/sunshineplan/password"
	"github.com/sunshineplan/utils/cache"
)

var userCache = cache.New[any, user](true)

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
	var ok bool
	if usr, ok = userCache.Get(id); ok {
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
	if err = accountClient.FindOne(filter, nil, &usr); err != nil {
		return
	}
	userCache.Set(usr.ID, usr, 24*time.Hour, nil)
	return
}

func authRequired(c *gin.Context) {
	if user, err := getUser(sessions.Default(c)); user.ID == "" || err == mongodb.ErrNoDocuments {
		c.AbortWithStatus(401)
	} else if user.ID != "" {
		c.Set("id", user.ID)
		c.Set("username", user.Username)
		c.Set("last", user.Last)
	} else {
		c.AbortWithStatus(500)
	}
}

type info struct {
	username any
	ip       string
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

	if password.IsMaxAttempts(info{login.Username, c.ClientIP()}) {
		c.JSON(200, gin.H{"status": 0, "message": fmt.Sprintf("Max retries exceeded (%d)", *maxRetry)})
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
		if err = password.CompareHashAndPassword(info{login.Username, c.ClientIP()}, user.Password, login.Password); err != nil {
			if errors.Is(err, password.ErrIncorrectPassword) {
				message = err.Error()
			} else if user.Password != login.Password {
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

	if password.IsMaxAttempts(info{username, c.ClientIP()}) {
		c.JSON(200, gin.H{"status": 0, "message": fmt.Sprintf("Max retries exceeded (%d)", *maxRetry), "error": 1})
		return
	}

	var data struct{ Password, Password1, Password2 string }
	if err := c.BindJSON(&data); err != nil {
		svc.Print(err)
		return
	}
	var err error
	if priv != nil {
		data.Password1, err = password.DecryptPKCS1v15(priv, data.Password1)
		if err != nil {
			c.String(400, "Bad Request")
			return
		}
		data.Password2, err = password.DecryptPKCS1v15(priv, data.Password2)
		if err != nil {
			c.String(400, "Bad Request")
			return
		}
	}

	var user user
	if err := accountClient.FindOne(mongodb.M{"_id": id.Interface()}, nil, &user); err != nil {
		svc.Print(err)
		c.AbortWithStatus(500)
		return
	}

	var message string
	var errorCode int
	if err = password.CompareHashAndPassword(info{username, c.ClientIP()}, user.Password, data.Password); err != nil {
		if errors.Is(err, password.ErrIncorrectPassword) {
			message = err.Error()
			errorCode = 1
		} else {
			svc.Print(err)
			c.String(500, "Internal Server Error")
			return
		}
	} else {
		if priv != nil {
			data.Password, _ = password.DecryptPKCS1v15(priv, data.Password)
		}
		switch {
		case data.Password1 != data.Password2:
			message = "confirm password doesn't match new password"
			errorCode = 2
		case data.Password1 == data.Password:
			message = "new password cannot be the same as old password"
			errorCode = 2
		case data.Password1 == "":
			message = "new password cannot be blank"
		}
	}

	if message == "" {
		newPassword, err := password.HashPassword(data.Password1)
		if err != nil {
			svc.Print(err)
			c.String(500, "Internal Server Error")
			return
		}
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
