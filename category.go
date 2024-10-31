package main

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sunshineplan/database/mongodb"
)

type category struct {
	ID       string `json:"_id,omitempty"`
	Category string `json:"category"`
	Count    int    `json:"count"`
}

func getCategory(c *gin.Context) {
	userID, _ := c.Get("id")
	var categories []category
	if err := bookmarkClient.Aggregate(
		[]mongodb.M{
			{"$match": mongodb.M{"user": userID}},
			{"$group": mongodb.M{"_id": "$category", "count": mongodb.M{"$sum": 1}}},
			{"$sort": mongodb.M{"_id": 1}},
		},
		&categories,
	); err != nil {
		svc.Println("Failed to query categories:", err)
		c.AbortWithStatus(500)
		return
	}
	res := []category{}
	var uncategorized category
	for _, i := range categories {
		i.Category = i.ID
		i.ID = ""
		if i.Category != "" {
			res = append(res, i)
		} else {
			uncategorized = i
		}
	}

	c.JSON(200, append(res, uncategorized))
}

func editCategory(c *gin.Context) {
	var data struct{ Old, New string }
	if err := c.BindJSON(&data); err != nil {
		svc.Print(err)
		return
	}
	data.New = strings.TrimSpace(data.New)

	userID, _ := c.Get("id")

	var message string
	var errorCode int
	exist, err := checkExist(mongodb.M{"user": userID, "category": data.New})
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
	case err != nil:
		svc.Println("Failed to get category:", err)
		c.AbortWithStatus(500)
		return
	case exist:
		message = fmt.Sprintf("Category %s is already existed.", data.New)
		errorCode = 1
	default:
		if _, err := bookmarkClient.UpdateMany(
			mongodb.M{"user": userID, "category": data.Old},
			mongodb.M{"$set": mongodb.M{"category": data.New}},
			nil,
		); err != nil {
			svc.Println("Failed to edit category:", err)
			c.AbortWithStatus(500)
			return
		}

		newLastModified(userID.(string), c)
		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}

func deleteCategory(c *gin.Context) {
	var data struct{ Category string }
	if err := c.BindJSON(&data); err != nil {
		svc.Print(err)
		return
	}

	userID, _ := c.Get("id")
	if _, err := bookmarkClient.UpdateMany(
		mongodb.M{"user": userID, "category": data.Category},
		mongodb.M{"$unset": mongodb.M{"category": 1}},
		nil,
	); err != nil {
		svc.Println("Failed to delete category:", err)
		c.AbortWithStatus(500)
		return
	}

	newLastModified(userID.(string), c)
	c.JSON(200, gin.H{"status": 1})
}
