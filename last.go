package main

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/database/mongodb"
)

var mu sync.Mutex

func checkLastModified(c *gin.Context) (string, bool) {
	v, _ := c.Cookie("last")
	last, _ := c.Get("last")
	return last.(string), v == last
}

func checkRequired(c *gin.Context) {
	if _, ok := checkLastModified(c); ok {
		c.Next()
	} else {
		c.AbortWithStatus(409)
	}
}

func newLastModified(id any, c *gin.Context) {
	last := strconv.FormatInt(time.Now().UnixNano(), 10)
	go updateLast(id, last)
	username, _ := c.Get("username")
	userCache.Swap(id, user{ID: id.(string), Username: username.(string), Last: last})
	c.SetCookie("last", last, 856400*365, "", "", false, true)
}

func updateLast(id any, last string) {
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
}
