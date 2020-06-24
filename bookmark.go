package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type bookmark struct {
	id       int
	name     string
	url      string
	category string
}

func getBookmark(c *gin.Context) {
	db, err := sql.Open("mysql", "user:password@/dbname")
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")

	var category gin.H
	var bookmarks []bookmark
	switch categoryID := c.Query("category"); categoryID {
	case "":
		category = gin.H{"id": -1, "name": "All Bookmarks"}
		rows, err := db.Query(`
SELECT bookmark.id, bookmark, url, category
FROM bookmark LEFT JOIN category ON category_id = category.id
WHERE bookmark.user_id = ? ORDER BY seq
`, userID)
		if err != nil {
			log.Println(err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var bookmark bookmark
			if err := rows.Scan(&bookmark.id, &bookmark.name, &bookmark.url, &bookmark.category); err != nil {
				log.Println(err)
				return
			}
			bookmarks = append(bookmarks, bookmark)
		}
	case "0":
		category = gin.H{"id": 0, "name": "Uncategorized"}
		rows, err := db.Query(`
SELECT id, bookmark, url FROM bookmark
WHERE category_id = 0 AND user_id = ? ORDER BY seq
`, userID)
		if err != nil {
			log.Println(err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var bookmark bookmark
			if err := rows.Scan(&bookmark.id, &bookmark.name, &bookmark.url); err != nil {
				log.Println(err)
				return
			}
			bookmarks = append(bookmarks, bookmark)
		}
	default:
		category = gin.H{"id": categoryID}
		var name string
		err = db.QueryRow("SELECT category FROM category WHERE id = ? AND user_id = ?",
			categoryID, userID).Scan(&name)
		category["name"] = name
		if err != nil {
			log.Println(err)
			c.String(403, "")
			return
		}
		rows, err := db.Query(`
SELECT id, bookmark, url FROM bookmark
WHERE category_id = ? AND user_id = ? ORDER BY seq
`, categoryID, userID)
		if err != nil {
			log.Println(err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var bookmark bookmark
			if err := rows.Scan(&bookmark.id, &bookmark.name, &bookmark.url); err != nil {
				log.Println(err)
				return
			}
			bookmarks = append(bookmarks, bookmark)
		}
		for _, b := range bookmarks {
			b.category = category["name"].(string)
		}
		c.HTML(200, "bookmark/index.html", gin.H{"category": category, "bookmarks": bookmarks})
	}
}

func addBookmark(c *gin.Context) {
	db, err := sql.Open("mysql", "user:password@/dbname")
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")
	categoryID := c.Query("category_id")

	var category string
	var categories []string
	if categoryID != "" {
		db.QueryRow("SELECT category FROM category WHERE id = ? AND user_id = ?", categoryID, userID).Scan(&category)
	} else {
		category = ""
	}
	rows, err := db.Query("SELECT category FROM category WHERE user_id = ? ORDER BY category", userID)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var categoryName string
		if err := rows.Scan(&categoryName); err != nil {
			log.Println(err)
			return
		}
		categories = append(categories, categoryName)
	}
	c.HTML(200, "bookmark/bookmark.html", gin.H{"id": 0, "bookmark": gin.H{"category": category}, "categories": categories})
}

func doAddBookmark(c *gin.Context) {
	db, err := sql.Open("mysql", "user:password@/dbname")
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")
	category := strings.TrimSpace(c.PostForm("category"))
	bookmark := strings.TrimSpace(c.PostForm("bookmark"))
	url := strings.TrimSpace(c.PostForm("url"))
	categoryID := getCategoryID(category, userID.(string))

	var exist, message string
	var errorCode int
	if bookmark == "" {
		message = "Bookmark name is empty."
		errorCode = 1
	} else if err = db.QueryRow("SELECT id FROM bookmark WHERE bookmark = ? AND user_id = ?", bookmark, userID).Scan(&exist); err != nil {
		message = fmt.Sprintf("Bookmark name %s is already existed.", bookmark)
		errorCode = 1
	} else if err = db.QueryRow("SELECT id FROM bookmark WHERE url = ? AND user_id = ?", url, userID).Scan(&exist); err != nil {
		message = fmt.Sprintf("Bookmark url %s is already existed.", url)
		errorCode = 2
	} else if categoryID == "" {
		message = "Category name exceeded length limit."
		errorCode = 3
	} else {
		db.Exec("INSERT INTO bookmark (bookmark, url, user_id, category_id) VALUES (?, ?, ?, ?)", bookmark, url, userID, categoryID)
		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}

func editBookmark(c *gin.Context) {
	db, err := sql.Open("mysql", "user:password@/dbname")
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")
	id := c.Param("id")

	var bookmark bookmark
	err = db.QueryRow(`
SELECT bookmark, url, category FROM bookmark
LEFT JOIN category ON category_id = category.id
WHERE bookmark.id = ? AND bookmark.user_id = ?
`, id, userID).Scan(&bookmark)
	if err != nil {
		c.String(403, "")
		return
	}

	var categories []string
	rows, err := db.Query("SELECT category FROM category WHERE user_id = ? ORDER BY category", userID)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			log.Println(err)
			return
		}
		categories = append(categories, category)
	}
	c.HTML(200, "bookmark/bookmark.html", gin.H{"id": id, "bookmark": bookmark, "categories": categories})
}

func doEditBookmark(c *gin.Context) {
	db, err := sql.Open("mysql", "user:password@/dbname")
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")
	id := c.Param("id")

	var old bookmark
	err = db.QueryRow(`
SELECT bookmark, url, category FROM bookmark
LEFT JOIN category ON category_id = category.id
WHERE bookmark.id = ? AND bookmark.user_id = ?
`, id, userID).Scan(&old)
	if err != nil {
		c.String(403, "")
		return
	}
	bookmark := strings.TrimSpace(c.PostForm("bookmark"))
	url := strings.TrimSpace(c.PostForm("url"))
	category := strings.TrimSpace(c.PostForm("category"))
	categoryID := getCategoryID(category, userID.(string))

	var exist, message string
	var errorCode int
	if bookmark == "" {
		message = "Bookmark name is empty."
		errorCode = 1
	} else if old.name == bookmark && old.url == url && old.category == category {
		message = "New bookmark is same as old bookmark."
	} else if err = db.QueryRow("SELECT id FROM bookmark WHERE bookmark = ? AND id != ? AND user_id = ?", bookmark, id, userID).Scan(&exist); err != nil {
		message = fmt.Sprintf("Bookmark name %s is already existed.", bookmark)
		errorCode = 1
	} else if err = db.QueryRow("SELECT id FROM bookmark WHERE url = ? AND id != ? AND user_id = ?", url, id, userID).Scan(&exist); err != nil {
		message = fmt.Sprintf("Bookmark url %s is already existed.", url)
		errorCode = 2
	} else if categoryID == "" {
		message = "Category name exceeded length limit."
		errorCode = 3
	} else {
		db.Exec("UPDATE bookmark SET bookmark = ?, url = ?, category_id = ? WHERE id = ? AND user_id = ?", bookmark, url, categoryID, id, userID)
		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}

func doDeleteBookmark(c *gin.Context) {
	db, err := sql.Open("mysql", "user:password@/dbname")
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")
	id := c.Param("id")

	db.Exec("DELETE FROM bookmark WHERE id = ? and user_id = ?", id, userID)
	c.JSON(200, gin.H{"status": 1})
}

func reorder(c *gin.Context) {
	db, err := sql.Open("mysql", "user:password@/dbname")
	if err != nil {
		log.Println(err)
		c.String(500, "")
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")

	orig := c.Query("orig")
	dest := c.Query("dest")
	refer := c.Query("refer")

	var origSeq, destSeq int

	db.QueryRow("SELECT seq FROM bookmark WHERE bookmark = ? AND user_id = ?", orig, userID).Scan(&origSeq)
	if dest != "#TOP_POSITION#" {
		err = db.QueryRow("SELECT seq FROM bookmark WHERE bookmark = ? AND user_id = ?", dest, userID).Scan(&destSeq)
	} else {
		err = db.QueryRow("SELECT seq FROM bookmark WHERE bookmark = ? AND user_id = ?", refer, userID).Scan(&destSeq)
		destSeq--
	}
	if err != nil {
		log.Println(err)
		c.String(500, "")
		return
	}

	if origSeq > destSeq {
		destSeq++
		_, err = db.Exec("UPDATE bookmark SET seq = seq+1 WHERE seq >= ? AND user_id = ? AND seq < ?", destSeq, userID, origSeq)
	} else {
		_, err = db.Exec("UPDATE bookmark SET seq = seq-1 WHERE seq <= ? AND user_id = ? AND seq > ?", destSeq, userID, origSeq)
	}
	if err != nil {
		log.Println(err)
		c.String(500, "")
		return
	}
	_, err = db.Exec("UPDATE bookmark SET seq = ? WHERE bookmark = ? AND user_id = ?", destSeq, orig, userID)
	if err != nil {
		log.Println(err)
		c.String(500, "")
		return
	}
	c.String(200, "1")
}
