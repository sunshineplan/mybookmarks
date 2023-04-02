package main

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	lastModified = make(map[any]int64)
	mu           sync.Mutex
)

func checkLastModified(id any, c *gin.Context) (string, bool) {
	last, ok := lastModified[id]
	if !ok {
		last = time.Now().UnixNano()
		lastModified[id] = last
	}
	v, _ := c.Cookie("last")
	if last := strconv.FormatInt(last, 10); v == last {
		return last, true
	} else {
		return last, false
	}
}

func checkRequired(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	userID, _ := c.Get("id")
	if last, ok := checkLastModified(userID, c); ok {
		c.Set("last", last)
		c.Next()
	} else {
		c.AbortWithStatus(409)
	}
}

func newLastModified(id any, c *gin.Context) {
	last := time.Now().UnixNano()
	lastModified[id] = last
	c.SetCookie("last", strconv.FormatInt(last, 10), 856400*365, "", "", false, true)
}
