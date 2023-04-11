package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/database/mongodb"
)

type bookmark struct {
	ID       string `json:"id"`
	ObjectID string `json:"_id,omitempty" bson:"_id"`
	Bookmark string `json:"bookmark"`
	URL      string `json:"url"`
	Category string `json:"category"`
	Seq      int    `json:"seq"`
}

func getBookmark(c *gin.Context) {
	var r struct{ Start int64 }
	c.ShouldBindJSON(&r)
	userID, _ := c.Get("id")
	bookmarks := []*bookmark{}
	if err := bookmarkClient.Find(
		mongodb.M{"user": userID}, &mongodb.FindOpt{Sort: mongodb.M{"seq": 1}, Skip: r.Start, Limit: 50}, &bookmarks,
	); err != nil {
		svc.Println("Failed to query bookmarks:", err)
		c.AbortWithStatus(500)
		return
	}
	for _, i := range bookmarks {
		i.ID = i.ObjectID
		i.ObjectID = ""
	}
	c.JSON(200, bookmarks)
}

func addBookmark(c *gin.Context) {
	var data bookmark
	if err := c.BindJSON(&data); err != nil {
		svc.Print(err)
		return
	}

	if data.Bookmark == "" {
		c.JSON(200, gin.H{"status": 0, "message": "Bookmark name is empty.", "error": 1})
		return
	} else if len(data.Category) > 15 {
		c.JSON(200, gin.H{"status": 0, "message": "Category name exceeded length limit.", "error": 3})
		return
	} else if data.Category == "All Bookmarks" || data.Category == "Uncategorized" {
		c.JSON(200, gin.H{"status": 0, "message": "Category name is not allow.", "error": 3})
		return
	}

	userID, _ := c.Get("id")
	ec := make(chan error, 2)
	var exist1, exist2 bool
	go func() {
		var err error
		exist1, err = checkExist(mongodb.M{"bookmark": data.Bookmark, "user": userID})
		ec <- err
	}()
	go func() {
		var err error
		exist2, err = checkExist(mongodb.M{"url": data.URL, "user": userID})
		ec <- err
	}()
	for i := 0; i < 2; i++ {
		if err := <-ec; err != nil {
			svc.Println("Failed to get bookmark:", err)
			c.AbortWithStatus(500)
			return
		}
	}

	var message string
	var errorCode int
	switch {
	case exist1:
		message = fmt.Sprintf("Bookmark name %s is already existed.", data.Bookmark)
		errorCode = 1
	case exist2:
		message = fmt.Sprintf("Bookmark url %s is already existed.", data.URL)
		errorCode = 2
	default:
		var bookmarks []bookmark
		if err := bookmarkClient.Find(
			mongodb.M{"user": userID}, &mongodb.FindOpt{Sort: mongodb.M{"seq": -1}, Limit: 1}, &bookmarks,
		); err != nil {
			svc.Println("Failed to get bookmarks:", err)
			c.AbortWithStatus(500)
			return
		}

		var seq int
		if len(bookmarks) == 0 {
			seq = 1
		} else {
			seq = bookmarks[0].Seq + 1
		}

		doc := struct {
			Bookmark string `json:"bookmark" bson:"bookmark"`
			URL      string `json:"url" bson:"url"`
			User     any    `json:"user" bson:"user"`
			Seq      int    `json:"seq" bson:"seq"`
			Created  any    `json:"created" bson:"created"`
			Category string `json:"category,omitempty" bson:"category,omitempty"`
		}{
			Bookmark: data.Bookmark,
			URL:      data.URL,
			User:     userID,
			Seq:      seq,
			Created:  bookmarkClient.Date(time.Now()).Interface(),
			Category: data.Category,
		}

		insertedID, err := bookmarkClient.InsertOne(doc)
		if err != nil {
			svc.Println("Failed to add bookmark:", err)
			c.AbortWithStatus(500)
			return
		}

		newLastModified(userID, c)
		c.JSON(200, gin.H{"status": 1, "id": insertedID.(mongodb.ObjectID).Hex(), "seq": seq})
		return
	}

	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}

func editBookmark(c *gin.Context) {
	id, err := bookmarkClient.ObjectID(c.Param("id"))
	if err != nil {
		svc.Print(err)
		c.AbortWithStatus(400)
		return
	}

	var new bookmark
	if err := c.BindJSON(&new); err != nil {
		svc.Print(err)
		return
	}

	if new.Bookmark == "" {
		c.JSON(200, gin.H{"status": 0, "message": "Bookmark name is empty.", "error": 1})
		return
	} else if len(new.Category) > 15 {
		c.JSON(200, gin.H{"status": 0, "message": "Category name exceeded length limit.", "error": 3})
		return
	} else if new.Category == "All Bookmarks" || new.Category == "Uncategorized" {
		c.JSON(200, gin.H{"status": 0, "message": "Category name is not allow.", "error": 3})
		return
	}

	userID, _ := c.Get("id")
	ec := make(chan error, 3)
	var old bookmark
	var exist1, exist2 bool
	go func() {
		ec <- bookmarkClient.FindOne(mongodb.M{"_id": id.Interface(), "user": userID}, nil, &old)
	}()
	go func() {
		var err error
		exist1, err = checkExist(mongodb.M{"_id": mongodb.M{"$ne": id.Interface()}, "bookmark": new.Bookmark, "user": userID})
		ec <- err
	}()
	go func() {
		var err error
		exist2, err = checkExist(mongodb.M{"_id": mongodb.M{"$ne": id.Interface()}, "url": new.URL, "user": userID})
		ec <- err
	}()
	for i := 0; i < 3; i++ {
		if err := <-ec; err != nil {
			svc.Println("Failed to get bookmark:", err)
			c.AbortWithStatus(500)
			return
		}
	}

	var message string
	var errorCode int
	switch {
	case old == new:
		message = "New bookmark is same as old bookmark."
	case exist1:
		message = fmt.Sprintf("Bookmark name %s is already existed.", new.Bookmark)
		errorCode = 1
	case exist2:
		message = fmt.Sprintf("Bookmark url %s is already existed.", new.URL)
		errorCode = 2
	default:
		var update mongodb.M
		if new.Category == "" {
			update = mongodb.M{"$set": mongodb.M{"bookmark": new.Bookmark, "url": new.URL}, "$unset": mongodb.M{"category": ""}}
		} else {
			update = mongodb.M{"$set": mongodb.M{"bookmark": new.Bookmark, "url": new.URL, "category": new.Category}}
		}
		if _, err := bookmarkClient.UpdateOne(mongodb.M{"_id": id.Interface()}, update, nil); err != nil {
			svc.Println("Failed to edit bookmark:", err)
			c.AbortWithStatus(500)
			return
		}

		newLastModified(userID, c)
		c.JSON(200, gin.H{"status": 1})
		return
	}

	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}

func deleteBookmark(c *gin.Context) {
	id, err := bookmarkClient.ObjectID(c.Param("id"))
	if err != nil {
		svc.Print(err)
		c.AbortWithStatus(400)
		return
	}
	userID, _ := c.Get("id")
	var bookmark bookmark
	if err := bookmarkClient.FindOneAndDelete(mongodb.M{"_id": id.Interface()}, nil, &bookmark); err != nil {
		if err == mongodb.ErrNoDocuments {
			c.AbortWithStatus(403)
			return
		}
		svc.Println("Failed to delete bookmark:", err)
		c.AbortWithStatus(500)
		return
	}

	if _, err := bookmarkClient.UpdateMany(
		mongodb.M{"user": userID, "seq": mongodb.M{"$gt": bookmark.Seq}},
		mongodb.M{"$inc": mongodb.M{"seq": -1}},
		nil,
	); err != nil {
		svc.Println("Failed to reorder after delete bookmark:", err)
		c.AbortWithStatus(500)
		return
	}

	newLastModified(userID, c)
	c.JSON(200, gin.H{"status": 1})
}

func reorder(c *gin.Context) {
	var data struct{ Orig, Dest string }
	if err := c.BindJSON(&data); err != nil {
		svc.Print(err)
		return
	}

	origID, err := bookmarkClient.ObjectID(data.Orig)
	if err != nil {
		svc.Print(err)
		c.AbortWithStatus(400)
		return
	}
	destID, err := bookmarkClient.ObjectID(data.Dest)
	if err != nil {
		svc.Print(err)
		c.AbortWithStatus(400)
		return
	}

	userID, _ := c.Get("id")
	for _, id := range []mongodb.ObjectID{origID, destID} {
		if exist, err := checkExist(mongodb.M{"_id": id.Interface(), "user": userID}); err != nil {
			svc.Println("Failed to get bookmark:", err)
			c.AbortWithStatus(500)
			return
		} else if !exist {
			c.AbortWithStatus(403)
			return
		}
	}

	if err := reorderBookmark(userID, origID, destID); err != nil {
		svc.Println("Failed to reorder bookmark:", err)
		c.AbortWithStatus(500)
		return
	}
	c.String(200, "1")
}
