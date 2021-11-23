package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/database/mongodb/api"
)

type category struct {
	ID       string `json:"_id,omitempty"`
	Category string `json:"category"`
	Count    int    `json:"count"`
}

func getCategory(userID interface{}) (categories []category, err error) {
	categories = []category{}
	if userID == nil {
		return
	}

	if err = bookmarkClient.Aggregate([]api.M{
		{"$match": api.M{"user": userID, "category": api.M{"$exists": true}}},
		{"$group": api.M{"_id": "$category", "count": api.M{"$sum": 1}}},
		{"$sort": api.M{"_id": 1}},
	},
		&categories,
	); err != nil {
		log.Println("Failed to query categories:", err)
	}
	for i := range categories {
		categories[i].Category = categories[i].ID
		categories[i].ID = ""
	}

	return
}

func editCategory(c *gin.Context) {
	var data struct{ Old, New string }
	if err := c.BindJSON(&data); err != nil {
		log.Print(err)
		c.String(400, "")
		return
	}
	data.New = strings.TrimSpace(data.New)

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
	case data.New == "All Bookmarks" || data.New == "Uncategorized":
		message = "New category name is not allow."
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
		if _, err := bookmarkClient.UpdateMany(
			api.M{"user": userID, "category": data.Old},
			api.M{"$set": api.M{"category": data.New}},
			nil,
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

	if _, err := bookmarkClient.UpdateMany(
		api.M{"user": userID, "category": data.Category},
		api.M{"$unset": api.M{"category": 1}},
		nil,
	); err != nil {
		log.Println("Failed to delete category:", err)
		c.String(500, "")
		return
	}

	c.JSON(200, gin.H{"status": 1})
}
