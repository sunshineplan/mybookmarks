package main

import (
	"context"
	"errors"
	"fmt"
	"log"
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

func checkBookmark(objecdID primitive.ObjectID, userID interface{}) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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
		log.Print(err)
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
		log.Print(err)
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
	} else if data.Category == "All Bookmarks" || data.Category == "Uncategorized" {
		c.JSON(200, gin.H{"status": 0, "message": "Category name is not allow.", "error": 3})
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

		document := bson.D{
			{Key: "bookmark", Value: data.Bookmark},
			{Key: "url", Value: data.URL},
			{Key: "user", Value: userID},
			{Key: "seq", Value: seq},
			{Key: "created", Value: time.Now()},
		}
		if data.Category != "" {
			document = append(document, bson.E{Key: "category", Value: data.Category})
		}

		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		res, err := collBookmark.InsertOne(ctx, document)
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
	objectID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	}

	var new bookmark
	if err := c.BindJSON(&new); err != nil {
		log.Print(err)
		c.String(400, "")
		return
	}

	userID, _, err := getUser(c)
	if err != nil {
		log.Print(err)
		c.String(500, "")
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

	ec := make(chan error, 3)
	var old bookmark
	var exist1, exist2 bool
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error
		err = collBookmark.FindOne(
			ctx, bson.M{"_id": objectID, "user": userID}).Decode(&old)
		if err == mongo.ErrNoDocuments {
			err = errors.New("Bookmark not found")
		}
		ec <- err
	}()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error
		if err = collBookmark.FindOne(
			ctx, bson.M{"_id": bson.M{"$ne": objectID}, "bookmark": new.Bookmark, "user": userID},
		).Err(); err == nil {
			exist1 = true
		}
		ec <- err
	}()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error
		if err = collBookmark.FindOne(
			ctx, bson.M{"_id": bson.M{"$ne": objectID}, "url": new.URL, "user": userID},
		).Err(); err == nil {
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
	case old == new:
		message = "New bookmark is same as old bookmark."
	case exist1:
		message = fmt.Sprintf("Bookmark name %s is already existed.", new.Bookmark)
		errorCode = 1
	case exist2:
		message = fmt.Sprintf("Bookmark url %s is already existed.", new.URL)
		errorCode = 2
	default:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var update bson.M
		if new.Category == "" {
			update = bson.M{"$set": bson.M{"bookmark": new.Bookmark, "url": new.URL}}
		} else {
			update = bson.M{"$set": bson.M{"bookmark": new.Bookmark, "url": new.URL, "category": new.Category}}
		}
		if _, err := collBookmark.UpdateOne(ctx, bson.M{"_id": objectID}, update); err != nil {
			log.Println("Failed to edit bookmark:", err)
			c.String(500, "")
			return
		}

		c.JSON(200, gin.H{"status": 1})
		return
	}

	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}

func deleteBookmark(c *gin.Context) {
	objectID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	}

	userID, _, err := getUser(c)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	} else if checkBookmark(objectID, userID) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var bookmark bookmark
		if err := collBookmark.FindOneAndDelete(ctx, bson.M{"_id": objectID}).Decode(&bookmark); err != nil {
			log.Println("Failed to delete bookmark:", err)
			c.String(500, "")
			return
		}

		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if _, err := collBookmark.UpdateMany(ctx,
			bson.M{"user": userID, "seq": bson.M{"$gt": bookmark.Seq}},
			bson.M{"$inc": bson.M{"seq": -1}},
		); err != nil {
			log.Println("Failed to reorder after delete bookmark:", err)
			c.String(500, "")
			return
		}

		c.JSON(200, gin.H{"status": 1})
		return
	}

	c.String(403, "")
}

func reorder(c *gin.Context) {
	var data struct{ Orig, Dest string }
	if err := c.BindJSON(&data); err != nil {
		log.Print(err)
		c.String(400, "")
		return
	}

	orig, err := primitive.ObjectIDFromHex(data.Orig)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	}

	dest, err := primitive.ObjectIDFromHex(data.Dest)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	}

	userID, _, err := getUser(c)
	if err != nil {
		log.Print(err)
		c.String(500, "")
		return
	} else if !checkBookmark(orig, userID) || !checkBookmark(dest, userID) {
		c.String(403, "")
		return
	}

	if err := reorderBookmark(userID, orig, dest); err != nil {
		log.Println("Failed to reorder bookmark:", err)
		c.String(500, "")
		return
	}
	c.String(200, "1")
}
