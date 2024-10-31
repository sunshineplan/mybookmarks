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

func newLastModified(id string, c *gin.Context) {
	last := strconv.FormatInt(time.Now().UnixNano(), 10)
	go updateLast(id, last)
	username, _ := c.Get("username")
	oid, _ := mongodb.OIDFromHex(id)
	userCache.Swap(id, user{ID: oid, Username: username.(string), Last: last})
	c.SetCookie("last", last, 856400*365, "", "", false, false)
}

func updateLast(id string, last string) {
	mu.Lock()
	defer mu.Unlock()
	objectID, _ := accountClient.ObjectID(id)
	if _, err := accountClient.UpdateOne(
		mongodb.M{"_id": objectID},
		mongodb.M{"$set": mongodb.M{"last": last}},
		nil,
	); err != nil {
		svc.Print(err)
	}
}
