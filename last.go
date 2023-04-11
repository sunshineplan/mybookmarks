package main

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/database/mongodb"
)

var mu sync.Mutex

func checkLastModified(id any, c *gin.Context) (string, bool) {
	v, _ := c.Cookie("last")
	last, _ := c.Get("last")
	return last.(string), v == last
}

func checkRequired(c *gin.Context) {
	id, _ := c.Get("id")
	if _, ok := checkLastModified(id, c); ok {
		c.Next()
	} else {
		c.AbortWithStatus(409)
	}
}

func newLastModified(id any, c *gin.Context) {
	last := time.Now().UnixNano()
	go func() {
		mu.Lock()
		defer mu.Unlock()
		objectID, _ := accountClient.ObjectID(id.(string))
		if _, err := accountClient.UpdateOne(
			mongodb.M{"_id": objectID.Interface()},
			mongodb.M{"$set": mongodb.M{"last": last}},
			nil,
		); err != nil {
			svc.Print(err)
		}
	}()
	c.SetCookie("last", strconv.FormatInt(last, 10), 856400*365, "", "", false, true)
}
