package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type bookmark struct {
	ID       string             `json:"id"`
	ObjectID primitive.ObjectID `json:"-" bson:"_id"`
	Bookmark string             `json:"bookmark"`
	URL      string             `json:"url"`
	Category string             `json:"category"`
	Seq      int                `json:"-"`
}

func checkBookmark(id, userID interface{}) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objecdID, err := primitive.ObjectIDFromHex(id.(string))
	if err != nil {
		log.Print(err)
		return false
	}

	if err := collBookmark.FindOne(ctx, bson.M{"_id": objecdID, "user": userID}).Err(); err != nil {
		if err != mongo.ErrNoDocuments {
			log.Print(err)
		}
		return false
	}

	return true
}

func getBookmark(userID interface{}) (bookmarks []bookmark, total int64, err error) {
	bookmarks = []bookmark{}
	if userID == nil {
		return
	}

	ec := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error
		total, err = collBookmark.CountDocuments(ctx, bson.M{"user": userID})
		ec <- err
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var cursor *mongo.Cursor
	cursor, err = collBookmark.Find(
		ctx, bson.M{"user": userID}, options.Find().SetSort(bson.M{"seq": 1}).SetLimit(50))
	if err != nil {
		log.Println("Failed to query bookmarks:", err)
		return
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = cursor.All(ctx, &bookmarks); err != nil {
		log.Println("Failed to get bookmarks:", err)
		return
	}

	err = <-ec

	return
}

func moreBookmark(c *gin.Context) {
	var r struct{ Start int64 }
	if err := c.BindJSON(&r); err != nil {
		c.String(400, "")
		return
	}

	userID, _, err := getUser(c)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collBookmark.Find(
		ctx, bson.M{"user": userID}, options.Find().SetSort(bson.M{"seq": 1}).SetSkip(r.Start).SetLimit(50))
	if err != nil {
		log.Println("Failed to query bookmarks:", err)
		c.String(500, "")
		return
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	bookmarks := []bookmark{}
	if err := cursor.All(ctx, &bookmarks); err != nil {
		log.Println("Failed to get bookmarks:", err)
		c.String(500, "")
		return
	}

	c.JSON(200, bookmarks)
}

func addBookmark(c *gin.Context) {
	var data bookmark
	if err := c.BindJSON(&data); err != nil {
		c.String(400, "")
		return
	}

	userID, _, err := getUser(c)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	}

	if data.Bookmark == "" {
		c.JSON(200, gin.H{"status": 0, "message": "Bookmark name is empty.", "error": 1})
		return
	} else if len(data.Category) > 15 {
		c.JSON(200, gin.H{"status": 0, "message": "Category name exceeded length limit.", "error": 3})
		return
	}

	ec := make(chan error, 2)
	var exist1, exist2 bool
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error
		if err = collBookmark.FindOne(
			ctx, bson.M{"bookmark": data.Bookmark, "user": userID}).Err(); err == nil {
			exist1 = true
		}
		ec <- err
	}()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error
		if err = collBookmark.FindOne(
			ctx, bson.M{"url": data.URL, "user": userID}).Err(); err == nil {
			exist2 = true
		}
		ec <- err
	}()
	for i := 0; i < 2; i++ {
		if err := <-ec; err != nil {
			if err != mongo.ErrNoDocuments {
				log.Println("Failed to get category id:", err)
				c.String(500, "")
				return
			}
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
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cursor, err := collBookmark.Find(
			ctx, bson.M{"user": userID}, options.Find().SetSort(bson.M{"seq": -1}).SetLimit(1))
		if err != nil {
			log.Println("Failed to query bookmarks:", err)
			c.String(500, "")
			return
		}

		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var bookmarks []bookmark
		if err := cursor.All(ctx, &bookmarks); err != nil {
			log.Println("Failed to get bookmarks:", err)
			c.String(500, "")
			return
		}

		var seq int
		if len(bookmarks) == 0 {
			seq = 1
		} else {
			seq = bookmarks[0].Seq
		}

		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		res, err := collBookmark.InsertOne(ctx, bson.D{
			{"bookmark", data.Bookmark},
			{"category", data.Category},
			{"url", data.URL},
			{"user", userID},
			{"seq", seq},
		})
		if err != nil {
			log.Println("Failed to add bookmark:", err)
			c.String(500, "")
			return
		}

		objecdID, ok := res.InsertedID.(primitive.ObjectID)
		if !ok {
			log.Print("Failed to get last insert id.")
			c.String(500, "")
			return
		}

		c.JSON(200, gin.H{"status": 1, "id": objecdID.Hex()})
		return
	}

	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}

func editBookmark(c *gin.Context) {
	id := c.Param("id")

	var new bookmark
	if err := c.BindJSON(&new); err != nil {
		c.String(400, "")
		return
	}

	userID, _, err := getUser(c)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	}

	bc := make(chan error, 4)
	var old bookmark
	var exist1, exist2 string
	go func() {
		var oldCategory []byte
		err := db.QueryRow("SELECT bookmark, url, category FROM bookmarks WHERE bookmark_id = ? AND user_id = ?",
			id, userID).Scan(&old.Name, &old.URL, &oldCategory)
		old.Category = string(oldCategory)
		bc <- err
	}()
	go func() {
		db.QueryRow("SELECT id FROM bookmark WHERE bookmark = ? AND id != ? AND user_id = ?",
			new.Name, id, userID).Scan(&exist1)
		bc <- nil
	}()
	go func() {
		db.QueryRow("SELECT id FROM bookmark WHERE url = ? AND id != ? AND user_id = ?",
			new.URL, id, userID).Scan(&exist2)
		bc <- nil
	}()
	for i := 0; i < 4; i++ {
		if err := <-bc; err != nil {
			log.Println(err)
			c.String(500, "")
			return
		}
	}

	var message string
	var errorCode int
	switch {
	case new.Name == "":
		message = "Bookmark name is empty."
		errorCode = 1
	case old == new:
		message = "New bookmark is same as old bookmark."
	case exist1 != "":
		message = fmt.Sprintf("Bookmark name %s is already existed.", new.Name)
		errorCode = 1
	case exist2 != "":
		message = fmt.Sprintf("Bookmark url %s is already existed.", new.URL)
		errorCode = 2
	case categoryID == -1:
		message = "Category name exceeded length limit."
		errorCode = 3
	default:
		if _, err := db.Exec("UPDATE bookmark SET bookmark = ?, url = ?, category_id = ? WHERE id = ? AND user_id = ?",
			new.Name, new.URL, categoryID, id, userID); err != nil {
			log.Println("Failed to edit bookmark:", err)
			c.String(500, "")
			return
		}
		c.JSON(200, gin.H{"status": 1, "cid": categoryID})
		return
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}

func deleteBookmark(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println("Failed to get id param:", err)
		c.String(400, "")
		return
	}

	userID, _, err := getUser(c)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	} else if checkBookmark(id, userID) {
		if _, err := db.Exec("DELETE FROM bookmark WHERE id = ? and user_id = ?",
			id, userID); err != nil {
			log.Println("Failed to delete bookmark:", err)
			c.String(500, "")
			return
		}
		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.String(403, "")
}

func reorder(c *gin.Context) {
	var reorder struct{ Old, New int }
	if err := c.BindJSON(&reorder); err != nil {
		c.String(400, "")
		return
	}

	userID, _, err := getUser(c)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	} else if !checkBookmark(reorder.Old, userID) || !checkBookmark(reorder.New, userID) {
		c.String(403, "")
		return
	}

	if _, err := db.Exec("CALL reorder(?, ?, ?)",
		userID, reorder.New, reorder.Old); err != nil {
		log.Println("Failed to reorder bookmark:", err)
		c.String(500, "")
		return
	}
	c.String(200, "1")
}
