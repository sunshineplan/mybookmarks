package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type bookmark struct {
	ID       int
	Name     string
	URL      string
	Category string
}

func getBookmark(c *gin.Context) {
	db, err := getDB()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		c.String(503, "")
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")

	var category gin.H

	stmt := "SELECT %s FROM mybookmarks WHERE"

	var args []interface{}
	categoryID, err := strconv.Atoi(c.Query("category"))
	switch {
	case err != nil:
		category = gin.H{"id": -1, "name": "All Bookmarks"}
		stmt += " user_id = ?"
		args = append(args, userID)
	case categoryID == 0:
		category = gin.H{"id": 0, "name": "Uncategorized"}
		stmt += " category_id = 0 AND user_id = ?"
		args = append(args, userID)
	default:
		category = gin.H{"id": categoryID}
		stmt += " category_id = ? AND user_id = ?"
		args = append(args, categoryID)
		args = append(args, userID)
	}

	rows, err := db.Query(fmt.Sprintf(stmt, "bookmark_id, bookmark, url, category"), args...)
	if err != nil {
		log.Printf("Failed to get bookmarks by category: %v", err)
		c.String(500, "")
		return
	}
	defer rows.Close()
	var bookmarks []bookmark
	for rows.Next() {
		var bookmark bookmark
		var category []byte
		if err := rows.Scan(&bookmark.ID, &bookmark.Name, &bookmark.URL, &category); err != nil {
			log.Printf("Failed to scan bookmarks by category: %v", err)
			c.String(500, "")
			return
		}
		bookmark.Category = string(category)
		bookmarks = append(bookmarks, bookmark)
	}
	c.JSON(200, gin.H{"category": category, "bookmarks": bookmarks})
}

func addBookmark(c *gin.Context) {
	db, err := getDB()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		c.String(503, "")
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")

	var bookmark bookmark
	var categories []string
	categoryID, err := strconv.Atoi(c.Query("category"))
	if err == nil {
		db.QueryRow("SELECT category FROM category WHERE id = ? AND user_id = ?", categoryID, userID).Scan(&bookmark.Category)
	} else {
		bookmark.Category = ""
	}
	rows, err := db.Query("SELECT category FROM category WHERE user_id = ? ORDER BY category", userID)
	if err != nil {
		log.Printf("Failed to get categories name: %v", err)
		c.String(500, "")
		return
	}
	defer rows.Close()
	for rows.Next() {
		var categoryName string
		if err := rows.Scan(&categoryName); err != nil {
			log.Printf("Failed to scan category name: %v", err)
			c.String(500, "")
			return
		}
		categories = append(categories, categoryName)
	}
	c.HTML(200, "bookmark.html", gin.H{"id": 0, "bookmark": bookmark, "categories": categories})
}

func doAddBookmark(c *gin.Context) {
	db, err := getDB()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		c.String(503, "")
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")
	category := strings.TrimSpace(c.PostForm("category"))
	bookmark := strings.TrimSpace(c.PostForm("bookmark"))
	url := strings.TrimSpace(c.PostForm("url"))
	categoryID, err := getCategoryID(category, userID.(int), db)
	if err != nil {
		log.Printf("Failed to get category id: %v", err)
		c.String(500, "")
		return
	}

	var exist, message string
	var errorCode int
	if bookmark == "" {
		message = "Bookmark name is empty."
		errorCode = 1
	} else if err = db.QueryRow("SELECT id FROM bookmark WHERE bookmark = ? AND user_id = ?", bookmark, userID).Scan(&exist); err == nil {
		message = fmt.Sprintf("Bookmark name %s is already existed.", bookmark)
		errorCode = 1
	} else if err = db.QueryRow("SELECT id FROM bookmark WHERE url = ? AND user_id = ?", url, userID).Scan(&exist); err == nil {
		message = fmt.Sprintf("Bookmark url %s is already existed.", url)
		errorCode = 2
	} else if categoryID == -1 {
		message = "Category name exceeded length limit."
		errorCode = 3
	} else {
		_, err = db.Exec("INSERT INTO bookmark (bookmark, url, user_id, category_id) VALUES (?, ?, ?, ?)", bookmark, url, userID, categoryID)
		if err != nil {
			log.Printf("Failed to add bookmark: %v", err)
			c.String(500, "")
			return
		}
		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}

func editBookmark(c *gin.Context) {
	db, err := getDB()
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

	var bookmark bookmark
	var category []byte
	if err := db.QueryRow("SELECT bookmark, url, category FROM mybookmarks WHERE bookmark_id = ? AND user_id = ?",
		id, userID).Scan(&bookmark.Name, &bookmark.URL, &category); err != nil {
		log.Printf("Failed to scan bookmark: %v", err)
		c.String(500, "")
		return
	}
	bookmark.Category = string(category)

	var categories []string
	rows, err := db.Query("SELECT category FROM category WHERE user_id = ? ORDER BY category", userID)
	if err != nil {
		log.Printf("Failed to get categories: %v", err)
		c.String(500, "")
		return
	}
	defer rows.Close()
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			log.Printf("Failed to scan category: %v", err)
			c.String(500, "")
			return
		}
		categories = append(categories, category)
	}
	c.HTML(200, "bookmark.html", gin.H{"id": id, "bookmark": bookmark, "categories": categories})
}

func doEditBookmark(c *gin.Context) {
	db, err := getDB()
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

	var old bookmark
	var oldCategory []byte
	if err := db.QueryRow("SELECT bookmark, url, category FROM mybookmarks WHERE bookmark_id = ? AND user_id = ?",
		id, userID).Scan(&old.Name, &old.URL, &oldCategory); err != nil {
		log.Printf("Failed to scan bookmark: %v", err)
		c.String(500, "")
		return
	}
	old.Category = string(oldCategory)
	bookmark := strings.TrimSpace(c.PostForm("bookmark"))
	url := strings.TrimSpace(c.PostForm("url"))
	category := strings.TrimSpace(c.PostForm("category"))
	categoryID, err := getCategoryID(category, userID.(int), db)
	if err != nil {
		log.Printf("Failed to get category id: %v", err)
		c.String(500, "")
		return
	}

	var exist, message string
	var errorCode int
	if bookmark == "" {
		message = "Bookmark name is empty."
		errorCode = 1
	} else if old.Name == bookmark && old.URL == url && string(old.Category) == category {
		message = "New bookmark is same as old bookmark."
	} else if err = db.QueryRow("SELECT id FROM bookmark WHERE bookmark = ? AND id != ? AND user_id = ?", bookmark, id, userID).Scan(&exist); err == nil {
		message = fmt.Sprintf("Bookmark name %s is already existed.", bookmark)
		errorCode = 1
	} else if err = db.QueryRow("SELECT id FROM bookmark WHERE url = ? AND id != ? AND user_id = ?", url, id, userID).Scan(&exist); err == nil {
		message = fmt.Sprintf("Bookmark url %s is already existed.", url)
		errorCode = 2
	} else if categoryID == -1 {
		message = "Category name exceeded length limit."
		errorCode = 3
	} else {
		_, err = db.Exec("UPDATE bookmark SET bookmark = ?, url = ?, category_id = ? WHERE id = ? AND user_id = ?", bookmark, url, categoryID, id, userID)
		if err != nil {
			log.Printf("Failed to edit bookmark: %v", err)
			c.String(500, "")
			return
		}
		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}

func doDeleteBookmark(c *gin.Context) {
	db, err := getDB()
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

	_, err = db.Exec("DELETE FROM bookmark WHERE id = ? and user_id = ?", id, userID)
	if err != nil {
		log.Printf("Failed to delete bookmark: %v", err)
		c.String(500, "")
		return
	}
	c.JSON(200, gin.H{"status": 1})
}

func reorder(c *gin.Context) {
	db, err := getDB()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		c.String(503, "")
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")

	orig := c.PostForm("orig")
	dest := c.PostForm("dest")
	next := c.PostForm("next")

	var origSeq, destSeq int

	err = db.QueryRow("SELECT seq FROM seq WHERE bookmark_id = ? AND user_id = ?", orig, userID).Scan(&origSeq)
	if err != nil {
		log.Printf("Failed to scan orig seq: %v", err)
		c.String(500, "")
		return
	}
	if dest != "#TOP_POSITION#" {
		err = db.QueryRow("SELECT seq FROM seq WHERE bookmark_id = ? AND user_id = ?", dest, userID).Scan(&destSeq)
	} else {
		err = db.QueryRow("SELECT seq FROM seq WHERE bookmark_id = ? AND user_id = ?", next, userID).Scan(&destSeq)
		destSeq--
	}
	if err != nil {
		log.Printf("Failed to scan dest seq: %v", err)
		c.String(500, "")
		return
	}

	if origSeq > destSeq {
		destSeq++
		_, err = db.Exec("UPDATE seq SET seq = seq+1 WHERE seq >= ? AND user_id = ? AND seq < ?", destSeq, userID, origSeq)
	} else {
		_, err = db.Exec("UPDATE seq SET seq = seq-1 WHERE seq <= ? AND user_id = ? AND seq > ?", destSeq, userID, origSeq)
	}
	if err != nil {
		log.Printf("Failed to update other seq: %v", err)
		c.String(500, "")
		return
	}
	_, err = db.Exec("UPDATE seq SET seq = ? WHERE bookmark_id = ? AND user_id = ?", destSeq, orig, userID)
	if err != nil {
		log.Printf("Failed to update orig seq: %v", err)
		c.String(500, "")
		return
	}
	c.String(200, "1")
}
