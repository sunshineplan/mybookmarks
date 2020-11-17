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
	ID       int    `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Category string `json:"category"`
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

	stmt := "SELECT %s FROM bookmarks WHERE"

	var args []interface{}
	categoryID, err := strconv.Atoi(fmt.Sprintf("%v", c.PostForm("category")))
	switch {
	case err != nil, categoryID == -1:
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

	var total int
	bc := make(chan bool, 1)
	go func() {
		if err := db.QueryRow(fmt.Sprintf(stmt, "count(*)"), args...).Scan(&total); err != nil {
			log.Printf("Failed to get total count: %v", err)
			bc <- false
		}
		bc <- true
	}()

	start := c.PostForm("start")
	if start == "" {
		start = "0"
	}
	limit := fmt.Sprintf(" LIMIT %v, 30", start)
	rows, err := db.Query(fmt.Sprintf(stmt+limit, "bookmark_id, bookmark, url, category"), args...)
	if err != nil {
		log.Printf("Failed to get bookmarks: %v", err)
		c.String(500, "")
		return
	}
	defer rows.Close()
	var bookmarks []bookmark
	for rows.Next() {
		var bookmark bookmark
		var categoryByte []byte
		if err := rows.Scan(&bookmark.ID, &bookmark.Name, &bookmark.URL, &categoryByte); err != nil {
			log.Printf("Failed to scan bookmarks: %v", err)
			c.String(500, "")
			return
		}
		bookmark.Category = string(categoryByte)
		bookmarks = append(bookmarks, bookmark)
		if category["id"].(int) > 0 {
			category["name"] = string(categoryByte)
		}
	}
	if v := <-bc; !v {
		c.String(500, "")
		return
	}
	c.JSON(200, gin.H{"category": category, "bookmarks": bookmarks, "total": total})
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
		if _, err := db.Exec("INSERT INTO bookmark (bookmark, url, user_id, category_id) VALUES (?, ?, ?, ?)",
			bookmark, url, userID, categoryID); err != nil {
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

	var old bookmark
	var oldCategory []byte
	if err := db.QueryRow("SELECT bookmark, url, category FROM bookmarks WHERE bookmark_id = ? AND user_id = ?",
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
	} else if err := db.QueryRow("SELECT id FROM bookmark WHERE bookmark = ? AND id != ? AND user_id = ?",
		bookmark, id, userID).Scan(&exist); err == nil {
		message = fmt.Sprintf("Bookmark name %s is already existed.", bookmark)
		errorCode = 1
	} else if err := db.QueryRow("SELECT id FROM bookmark WHERE url = ? AND id != ? AND user_id = ?",
		url, id, userID).Scan(&exist); err == nil {
		message = fmt.Sprintf("Bookmark url %s is already existed.", url)
		errorCode = 2
	} else if categoryID == -1 {
		message = "Category name exceeded length limit."
		errorCode = 3
	} else {
		if _, err := db.Exec("UPDATE bookmark SET bookmark = ?, url = ?, category_id = ? WHERE id = ? AND user_id = ?",
			bookmark, url, categoryID, id, userID); err != nil {
			log.Printf("Failed to edit bookmark: %v", err)
			c.String(500, "")
			return
		}
		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}

func deleteBookmark(c *gin.Context) {
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

	if _, err := db.Exec("DELETE FROM bookmark WHERE id = ? and user_id = ?", id, userID); err != nil {
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

	if err := db.QueryRow("SELECT seq FROM seq WHERE bookmark_id = ? AND user_id = ?",
		orig, userID).Scan(&origSeq); err != nil {
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
	if _, err := db.Exec("UPDATE seq SET seq = ? WHERE bookmark_id = ? AND user_id = ?",
		destSeq, orig, userID); err != nil {
		log.Printf("Failed to update orig seq: %v", err)
		c.String(500, "")
		return
	}
	c.String(200, "1")
}
