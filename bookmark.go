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

type bookmark struct {
	ID       int
	Name     string
	URL      string
	Category []byte
}

func getBookmark(c *gin.Context) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println(err)
		c.String(503, "")
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")

	var category gin.H
	var bookmarks []bookmark
	categoryID, err := strconv.Atoi(c.Query("category"))
	switch {
	case err != nil:
		category = gin.H{"id": -1, "name": "All Bookmarks"}
		rows, err := db.Query(`
SELECT bookmark.id, bookmark, url, category
FROM bookmark LEFT JOIN category ON category_id = category.id
LEFT JOIN seq ON bookmark.user_id = seq.user_id AND bookmark.id = seq.bookmark_id
WHERE bookmark.user_id = ? ORDER BY seq
`, userID)
		if err != nil {
			log.Println(err)
			c.String(500, "")
			return
		}
		defer rows.Close()
		for rows.Next() {
			var bookmark bookmark
			if err := rows.Scan(&bookmark.ID, &bookmark.Name, &bookmark.URL, &bookmark.Category); err != nil {
				log.Println(err)
				c.String(500, "")
				return
			}
			bookmarks = append(bookmarks, bookmark)
		}
	case categoryID == 0:
		category = gin.H{"id": 0, "name": "Uncategorized"}
		rows, err := db.Query(`
SELECT id, bookmark, url FROM bookmark
LEFT JOIN seq ON bookmark.user_id = seq.user_id AND bookmark.id = seq.bookmark_id
WHERE category_id = 0 AND bookmark.user_id = ? ORDER BY seq
`, userID)
		if err != nil {
			log.Println(err)
			c.String(500, "")
			return
		}
		defer rows.Close()
		for rows.Next() {
			var bookmark bookmark
			if err := rows.Scan(&bookmark.ID, &bookmark.Name, &bookmark.URL); err != nil {
				log.Println(err)
				c.String(500, "")
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
			c.String(500, "")
			return
		}
		rows, err := db.Query(`
SELECT id, bookmark, url FROM bookmark
LEFT JOIN seq ON bookmark.user_id = seq.user_id AND bookmark.id = seq.bookmark_id
WHERE category_id = ? AND bookmark.user_id = ? ORDER BY seq
`, categoryID, userID)
		if err != nil {
			log.Println(err)
			c.String(500, "")
			return
		}
		defer rows.Close()
		for rows.Next() {
			var bookmark bookmark
			if err := rows.Scan(&bookmark.ID, &bookmark.Name, &bookmark.URL); err != nil {
				log.Println(err)
				c.String(500, "")
				return
			}
			bookmarks = append(bookmarks, bookmark)
		}
		for i := range bookmarks {
			bookmarks[i].Category = []byte(name)
		}
	}
	c.HTML(200, "index.html", gin.H{"category": category, "bookmarks": bookmarks})
}

func addBookmark(c *gin.Context) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println(err)
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
		bookmark.Category = []byte("")
	}
	rows, err := db.Query("SELECT category FROM category WHERE user_id = ? ORDER BY category", userID)
	if err != nil {
		log.Println(err)
		c.String(500, "")
		return
	}
	defer rows.Close()
	for rows.Next() {
		var categoryName string
		if err := rows.Scan(&categoryName); err != nil {
			log.Println(err)
			c.String(500, "")
			return
		}
		categories = append(categories, categoryName)
	}
	c.HTML(200, "bookmark.html", gin.H{"id": 0, "bookmark": bookmark, "categories": categories})
}

func doAddBookmark(c *gin.Context) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println(err)
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
		log.Println(err)
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
			log.Println(err)
			c.String(500, "")
			return
		}
		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}

func editBookmark(c *gin.Context) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println(err)
		c.String(503, "")
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println(err)
		c.String(400, "")
		return
	}

	var bookmark bookmark
	err = db.QueryRow(`
SELECT bookmark, url, category FROM bookmark
LEFT JOIN category ON category_id = category.id
WHERE bookmark.id = ? AND bookmark.user_id = ?
`, id, userID).Scan(&bookmark.Name, &bookmark.URL, &bookmark.Category)
	if err != nil {
		log.Println(err)
		c.String(500, "")
		return
	}

	var categories []string
	rows, err := db.Query("SELECT category FROM category WHERE user_id = ? ORDER BY category", userID)
	if err != nil {
		log.Println(err)
		c.String(500, "")
		return
	}
	defer rows.Close()
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			log.Println(err)
			c.String(500, "")
			return
		}
		categories = append(categories, category)
	}
	c.HTML(200, "bookmark.html", gin.H{"id": id, "bookmark": bookmark, "categories": categories})
}

func doEditBookmark(c *gin.Context) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println(err)
		c.String(503, "")
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println(err)
		c.String(400, "")
		return
	}

	var old bookmark
	err = db.QueryRow(`
SELECT bookmark, url, category FROM bookmark
LEFT JOIN category ON category_id = category.id
WHERE bookmark.id = ? AND bookmark.user_id = ?
`, id, userID).Scan(&old.Name, &old.URL, &old.Category)
	if err != nil {
		log.Println(err)
		c.String(500, "")
		return
	}
	bookmark := strings.TrimSpace(c.PostForm("bookmark"))
	url := strings.TrimSpace(c.PostForm("url"))
	category := strings.TrimSpace(c.PostForm("category"))
	categoryID, _ := getCategoryID(category, userID.(int), db)

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
			log.Println(err)
			c.String(500, "")
			return
		}
		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}

func doDeleteBookmark(c *gin.Context) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println(err)
		c.String(503, "")
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println(err)
		c.String(400, "")
		return
	}

	_, err = db.Exec("DELETE FROM bookmark WHERE id = ? and user_id = ?", id, userID)
	if err != nil {
		log.Println(err)
		c.String(500, "")
		return
	}
	c.JSON(200, gin.H{"status": 1})
}

func reorder(c *gin.Context) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println(err)
		c.String(500, "")
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
		log.Println(err)
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
		log.Println(err)
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
		log.Println(err)
		c.String(500, "")
		return
	}
	_, err = db.Exec("UPDATE seq SET seq = ? WHERE bookmark_id = ? AND user_id = ?", destSeq, orig, userID)
	if err != nil {
		log.Println(err)
		c.String(500, "")
		return
	}
	c.String(200, "1")
}
