package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

type category struct {
	ID    int    `json:"id"`
	Name  string `json:"category"`
	Count int    `json:"count"`
}

func checkCategory(categoryID, userID interface{}) bool {
	var exist string
	if err := db.QueryRow("SELECT category FROM category WHERE id = ? AND user_id = ?",
		categoryID, userID).Scan(&exist); err == nil {
		return true
	}
	return false
}

func getCategory(userID interface{}) (categories []category, err error) {
	categories = []category{}
	if userID == nil {
		return
	}

	var rows *sql.Rows
	rows, err = db.Query("SELECT id, category, count FROM categories WHERE user_id = ?", userID)
	if err != nil {
		log.Println("Failed to get categories:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var category category
		if err = rows.Scan(&category.ID, &category.Name, &category.Count); err != nil {
			log.Println("Failed to scan category:", err)
			return
		}
		categories = append(categories, category)
	}

	return
}

func getCategoryID(category string, userID int) (int, error) {
	if category != "" {
		var categoryID int
		err := db.QueryRow("SELECT id FROM category WHERE category = ? AND user_id = ?", category, userID).Scan(&categoryID)
		switch {
		case len(category) > 15:
			return -1, nil
		case err != nil:
			if err == sql.ErrNoRows {
				res, err := db.Exec("INSERT INTO category (category, user_id) VALUES (?, ?)", category, userID)
				if err != nil {
					log.Println("Failed to add category:", err)
					return 0, err
				}
				lastInsertID, err := res.LastInsertId()
				if err != nil {
					log.Println("Failed to get last insert id:", err)
					return 0, err
				}
				return int(lastInsertID), nil
			}
			return 0, err
		default:
			return categoryID, nil
		}
	} else {
		return 0, nil
	}
}

func addCategory(c *gin.Context) {
	var category category
	if err := c.BindJSON(&category); err != nil {
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
	switch {
	case category.Name == "":
		message = "Category name is empty."
	case len(category.Name) > 15:
		message = "Category name exceeded length limit."
	default:
		var exist string
		if err := db.QueryRow("SELECT id FROM category WHERE category = ? AND user_id = ?",
			category.Name, userID).Scan(&exist); err == nil {
			message = fmt.Sprintf("Category %s is already existed.", category.Name)
		} else {
			if err == sql.ErrNoRows {
				res, err := db.Exec("INSERT INTO category (category, user_id) VALUES (?, ?)",
					category.Name, userID)
				if err != nil {
					log.Println("Failed to add category:", err)
					c.String(500, "")
					return
				}
				id, err := res.LastInsertId()
				if err != nil {
					log.Println("Failed to get last insert id:", err)
					c.String(500, "")
					return
				}
				c.JSON(200, gin.H{"status": 1, "id": id})
				return
			}
			log.Println("Failed to scan category:", err)
			c.String(500, "")
			return
		}
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": 1})
}

func editCategory(c *gin.Context) {
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
	} else if !checkCategory(id, userID) {
		c.String(403, "")
		return
	}

	var category category
	if err := c.BindJSON(&category); err != nil {
		c.String(400, "")
		return
	}

	ec := make(chan error, 1)
	var exist string
	go func() {
		ec <- db.QueryRow("SELECT id FROM category WHERE category = ? AND user_id = ?",
			category.Name, userID).Scan(&exist)
	}()
	var oldCategory string
	if err := db.QueryRow("SELECT category FROM category WHERE id = ? AND user_id = ?",
		id, userID).Scan(&oldCategory); err != nil {
		log.Println("Failed to scan category:", err)
		c.String(500, "")
		return
	}
	err = <-ec

	var message string
	var errorCode int
	switch {
	case category.Name == "":
		message = "New category name is empty."
		errorCode = 1
	case oldCategory == category.Name:
		message = "New category is same as old category."
	case len(category.Name) > 15:
		message = "Category name exceeded length limit."
		errorCode = 1
	case err == nil:
		message = fmt.Sprintf("Category %s is already existed.", category.Name)
		errorCode = 1
	default:
		if _, err := db.Exec("UPDATE category SET category = ? WHERE id = ? AND user_id = ?",
			category.Name, id, userID); err != nil {
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
	} else if checkCategory(id, userID) {
		if _, err := db.Exec("CALL delete_category(?)", id); err != nil {
			log.Println("Failed to deleted category:", err)
			c.String(500, "")
			return
		}
		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.String(403, "")
}
