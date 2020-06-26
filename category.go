package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type category struct {
	id    int
	name  string
	count int
}

func getCategoryID(category, userID string) string {
	if category != "" {
		db, _ := sql.Open("mysql", dsn)
		defer db.Close()
		var categoryID string
		db.QueryRow("SELECT id FROM category WHERE category = ? AND user_id = ?", category, userID).Scan(&categoryID)
		switch {
		case len(category) > 15:
			return ""
		case categoryID != "0":
			return categoryID
		default:
			res, _ := db.Exec("INSERT INTO category (category, user_id) VALUES (?, ?)", category, userID)
			lastInsertID, _ := res.LastInsertId()
			return strconv.FormatInt(lastInsertID, 10)
		}
	} else {
		return "0"
	}
}

func getCategory(c *gin.Context) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")

	var total, uncategorized int
	var categories []category
	db.QueryRow("SELECT count(bookmark) num FROM bookmark WHERE user_id = ?", userID).Scan(&total)
	db.QueryRow(
		"SELECT count(bookmark) num FROM bookmark WHERE category_id = 0 AND user_id = ?", userID).Scan(&uncategorized)
	rows, err := db.Query(`
SELECT category.id, category, count(bookmark) num
FROM category LEFT JOIN bookmark ON category.id = category_id
WHERE category.user_id = ? GROUP BY category.id ORDER BY category
`, userID)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var category category
		if err := rows.Scan(&category.id, &category.name, &category.count); err != nil {
			log.Println(err)
			return
		}
		categories = append(categories, category)
	}
	c.JSON(200, gin.H{"total": total, "uncategorized": uncategorized, "categories": categories})
}

func addCategory(c *gin.Context) {
	c.HTML(200, "bookmark/category.html", gin.H{"id": 0})
}

func doAddCategory(c *gin.Context) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")

	var message string
	category := strings.TrimSpace(c.PostForm("category"))
	switch {
	case category == "":
		message = "Category name is empty."
	case len(category) > 15:
		message = "Category name exceeded length limit."
	default:
		var exist string
		err = db.QueryRow("SELECT id FROM category WHERE category = ? AND user_id = ?", category, userID).Scan(&exist)
		if err != nil {
			message = fmt.Sprintf("Category %s is already existed.", category)
		} else {
			db.Exec("INSERT INTO category (category, user_id) VALUES (?, ?)", category, userID)
			c.JSON(200, gin.H{"status": 1})
			return
		}
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": 1})
}

func editCategory(c *gin.Context) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")

	var category category
	err = db.QueryRow("SELECT * FROM category WHERE id = ? AND user_id = ?", userID).Scan(&category)
	if err != nil {
		c.String(403, "")
		return
	}
	id := c.Param("id")
	c.HTML(200, "bookmark/category.html", gin.H{"id": id, "category": category})
}

func doEditCategory(c *gin.Context) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")
	id := c.Param("id")

	var oldCategory string
	err = db.QueryRow("SELECT category FROM category WHERE id = ? AND user_id = ?", userID).Scan(&oldCategory)
	if err != nil {
		c.String(403, "")
		return
	}
	newCategory := strings.TrimSpace(c.PostForm("category"))
	var message string
	var errorCode int
	switch {
	case newCategory == "":
		message = "New category name is empty."
		errorCode = 1
	case oldCategory == newCategory:
		message = "New category is same as old category."
	case len(newCategory) > 15:
		message = "Category name exceeded length limit."
		errorCode = 1
	default:
		var exist string
		err = db.QueryRow("SELECT id FROM category WHERE category = ? AND user_id = ?", newCategory, userID).Scan(&exist)
		if err != nil {
			message = fmt.Sprintf("Category %s is already existed.", newCategory)
			errorCode = 1
		} else {
			db.Exec("UPDATE category SET category = ? WHERE id = ? AND user_id = ?", newCategory, id, userID)
			c.JSON(200, gin.H{"status": 1})
			return
		}
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}

func doDeleteCategory(c *gin.Context) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")
	id := c.Param("id")

	db.Exec("DELETE FROM category WHERE id = ? and user_id = ?", id, userID)
	db.Exec("UPDATE bookmark SET category_id = 0 WHERE category_id = ? and user_id = ?", id, userID)
	c.JSON(200, gin.H{"status": 1})
}
