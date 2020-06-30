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
	ID    int
	Name  string
	Count int
}

func getCategoryID(category string, userID int, db *sql.DB) (int, error) {
	if category != "" {
		var categoryID int
		err := db.QueryRow("SELECT id FROM category WHERE category = ? AND user_id = ?", category, userID).Scan(&categoryID)
		switch {
		case len(category) > 15:
			return -1, nil
		case err != nil:
			if strings.Contains(err.Error(), "no rows") {
				res, err := db.Exec("INSERT INTO category (category, user_id) VALUES (?, ?)", category, userID)
				if err != nil {
					log.Printf("Failed to add category: %v", err)
					return 0, err
				}
				lastInsertID, err := res.LastInsertId()
				if err != nil {
					log.Printf("Failed to get last insert id: %v", err)
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

func getCategory(c *gin.Context) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		c.String(503, "")
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")

	var total, uncategorized int
	var categories []category
	err = db.QueryRow("SELECT count(bookmark) num FROM bookmark WHERE user_id = ?", userID).Scan(&total)
	if err != nil {
		log.Printf("Failed to scan all bookmark count: %v", err)
		c.String(500, "")
		return
	}
	err = db.QueryRow("SELECT count(bookmark) num FROM bookmark WHERE category_id = 0 AND user_id = ?", userID).Scan(&uncategorized)
	if err != nil {
		log.Printf("Failed to scan uncategorized bookmark count: %v", err)
		c.String(500, "")
		return
	}
	rows, err := db.Query(`
SELECT category.id, category, count(bookmark) num
FROM category LEFT JOIN bookmark ON category.id = category_id
WHERE category.user_id = ? GROUP BY category.id ORDER BY category
`, userID)
	if err != nil {
		log.Printf("Failed to get categories: %v", err)
		c.String(500, "")
		return
	}
	defer rows.Close()
	for rows.Next() {
		var category category
		if err := rows.Scan(&category.ID, &category.Name, &category.Count); err != nil {
			log.Printf("Failed to scan category: %v", err)
			c.String(500, "")
			return
		}
		categories = append(categories, category)
	}
	c.JSON(200, gin.H{"total": total, "uncategorized": uncategorized, "categories": categories})
}

func doAddCategory(c *gin.Context) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		c.String(503, "")
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
		if err == nil {
			message = fmt.Sprintf("Category %s is already existed.", category)
		} else {
			if strings.Contains(err.Error(), "no rows") {
				_, err = db.Exec("INSERT INTO category (category, user_id) VALUES (?, ?)", category, userID)
				if err != nil {
					log.Printf("Failed to add category: %v", err)
					c.String(500, "")
					return
				}
				c.JSON(200, gin.H{"status": 1})
				return
			}
			log.Printf("Failed to scan category: %v", err)
			c.String(500, "")
			return
		}
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": 1})
}

func editCategory(c *gin.Context) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		c.String(503, "")
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Failed to get id param: %v", err)
		c.String(400, "")
		return
	}

	var category category
	err = db.QueryRow("SELECT id, category FROM category WHERE id = ? AND user_id = ?", id, userID).Scan(&category.ID, &category.Name)
	if err != nil {
		log.Printf("Failed to scan category: %v", err)
		c.String(500, "")
		return
	}
	c.HTML(200, "category.html", gin.H{"id": id, "category": category})
}

func doEditCategory(c *gin.Context) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		c.String(503, "")
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Failed to get id param: %v", err)
		c.String(400, "")
		return
	}

	var oldCategory string
	err = db.QueryRow("SELECT category FROM category WHERE id = ? AND user_id = ?", id, userID).Scan(&oldCategory)
	if err != nil {
		log.Printf("Failed to scan category: %v", err)
		c.String(500, "")
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
		if err == nil {
			message = fmt.Sprintf("Category %s is already existed.", newCategory)
			errorCode = 1
		} else {
			_, err = db.Exec("UPDATE category SET category = ? WHERE id = ? AND user_id = ?", newCategory, id, userID)
			if err != nil {
				log.Printf("Failed to edit category: %v", err)
				c.String(500, "")
				return
			}
			c.JSON(200, gin.H{"status": 1})
			return
		}
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}

func doDeleteCategory(c *gin.Context) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		c.String(503, "")
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Failed to get id param: %v", err)
		c.String(400, "")
		return
	}

	_, err = db.Exec("DELETE FROM category WHERE id = ? and user_id = ?", id, userID)
	if err != nil {
		log.Printf("Failed to delete category: %v", err)
		c.String(500, "")
		return
	}
	_, err = db.Exec("UPDATE bookmark SET category_id = 0 WHERE category_id = ? and user_id = ?", id, userID)
	if err != nil {
		log.Printf("Failed to remove deleted category for bookmark: %v", err)
		c.String(500, "")
		return
	}
	c.JSON(200, gin.H{"status": 1})
}
