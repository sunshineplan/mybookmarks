package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type category struct {
	Category string `json:"category" bson:"_id"`
	Count    int    `json:"count"`
}

func getCategory(userID interface{}) (categories []category, err error) {
	categories = []category{}
	if userID == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var cursor *mongo.Cursor
	cursor, err = collBookmark.Aggregate(ctx, []bson.M{
		{"$match": bson.M{"user": userID, "category": bson.M{"$exists": true}}},
		{"$group": bson.M{"_id": "$category", "count": bson.M{"$sum": 1}}},
		{"$sort": bson.M{"_id": 1}},
	})
	if err != nil {
		log.Println("Failed to query categories:", err)
		return
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = cursor.All(ctx, &categories); err != nil {
		log.Println("Failed to get categories:", err)
	}

	return
}

func editCategory(c *gin.Context) {
	var data struct{ Old, New string }
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

	var message string
	var errorCode int
	switch {
	case data.New == "":
		message = "New category name is empty."
		errorCode = 1
	case data.Old == data.New:
		message = "New category is same as old category."
	case len(data.New) > 15:
		message = "Category name exceeded length limit."
		errorCode = 1
	case err == nil:
		message = fmt.Sprintf("Category %s is already existed.", data.New)
		errorCode = 1
	default:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if _, err := collBookmark.UpdateMany(ctx,
			bson.M{"user": userID, "category": data.Old},
			bson.M{"$set": bson.M{"category": data.New}},
		); err != nil {
			log.Println("Failed to edit category:", err)
			c.String(500, "")
			return
		}

		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}

func deleteCategory(c *gin.Context) {
	var data struct{ Category string }
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := collBookmark.DeleteMany(
		ctx, bson.M{"user": userID, "category": data.Category}); err != nil {
		log.Println("Failed to deleted category:", err)
		c.String(500, "")
		return
	}

	c.JSON(200, gin.H{"status": 1})
	return
}
