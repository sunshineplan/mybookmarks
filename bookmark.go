package main

import (
	"fmt"
	"log"
	"strconv"

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
	var r struct{ Category, Start int }
	if err := c.BindJSON(&r); err != nil {
		c.String(400, "")
		return
	}

	db, err := getDB()
	if err != nil {
		log.Println("Failed to connect to database:", err)
		c.String(503, "")
		return
	}
	defer db.Close()

	session := sessions.Default(c)
	userID := session.Get("user_id")

	stmt := "SELECT %s FROM bookmarks WHERE"

	var args []interface{}
	switch r.Category {
	case -1:
		stmt += " user_id = ?"
		args = append(args, userID)
	case 0:
		stmt += " category_id = 0 AND user_id = ?"
		args = append(args, userID)
	default:
		stmt += " category_id = ? AND user_id = ?"
		args = append(args, r.Category)
		args = append(args, userID)
	}

	limit := fmt.Sprintf(" LIMIT %d, 30", r.Start)
	rows, err := db.Query(fmt.Sprintf(stmt+limit, "bookmark_id, bookmark, url, category"), args...)
	if err != nil {
		log.Println("Failed to get bookmarks:", err)
		c.String(500, "")
		return
	}
	defer rows.Close()
	bookmarks := []bookmark{}
	for rows.Next() {
		var bookmark bookmark
		var categoryByte []byte
		if err := rows.Scan(&bookmark.ID, &bookmark.Name, &bookmark.URL, &categoryByte); err != nil {
			log.Println("Failed to scan bookmarks:", err)
			c.String(500, "")
			return
		}
		bookmark.Category = string(categoryByte)
		bookmarks = append(bookmarks, bookmark)
	}
	c.JSON(200, bookmarks)
}

func addBookmark(c *gin.Context) {
	db, err := getDB()
	if err != nil {
		log.Println("Failed to connect to database:", err)
		c.String(503, "")
		return
	}
	defer db.Close()

	session := sessions.Default(c)
	userID := session.Get("user_id")

	var bookmark bookmark
	if err := c.BindJSON(&bookmark); err != nil {
		c.String(400, "")
		return
	}
	categoryID, err := getCategoryID(bookmark.Category, userID.(int), db)
	if err != nil {
		log.Println("Failed to get category id:", err)
		c.String(500, "")
		return
	}

	var exist, message string
	var errorCode int
	if bookmark.Name == "" {
		message = "Bookmark name is empty."
		errorCode = 1
	} else if err = db.QueryRow("SELECT id FROM bookmark WHERE bookmark = ? AND user_id = ?",
		bookmark.Name, userID).Scan(&exist); err == nil {
		message = fmt.Sprintf("Bookmark name %s is already existed.", bookmark.Name)
		errorCode = 1
	} else if err = db.QueryRow("SELECT id FROM bookmark WHERE url = ? AND user_id = ?",
		bookmark.URL, userID).Scan(&exist); err == nil {
		message = fmt.Sprintf("Bookmark url %s is already existed.", bookmark.URL)
		errorCode = 2
	} else if categoryID == -1 {
		message = "Category name exceeded length limit."
		errorCode = 3
	} else {
		if _, err := db.Exec("INSERT INTO bookmark (bookmark, url, user_id, category_id) VALUES (?, ?, ?, ?)",
			bookmark.Name, bookmark.URL, userID, categoryID); err != nil {
			log.Println("Failed to add bookmark:", err)
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
		log.Println("Failed to connect to database:", err)
		c.String(503, "")
		return
	}
	defer db.Close()

	session := sessions.Default(c)
	userID := session.Get("user_id")

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println("Failed to get id param:", err)
		c.String(400, "")
		return
	}

	var old bookmark
	var oldCategory []byte
	if err := db.QueryRow("SELECT bookmark, url, category FROM bookmarks WHERE bookmark_id = ? AND user_id = ?",
		id, userID).Scan(&old.Name, &old.URL, &oldCategory); err != nil {
		log.Println("Failed to scan bookmark:", err)
		c.String(500, "")
		return
	}
	old.Category = string(oldCategory)

	var bookmark bookmark
	if err := c.BindJSON(&bookmark); err != nil {
		c.String(400, "")
		return
	}
	categoryID, err := getCategoryID(bookmark.Category, userID.(int), db)
	if err != nil {
		log.Println("Failed to get category id:", err)
		c.String(500, "")
		return
	}

	var exist, message string
	var errorCode int
	if bookmark.Name == "" {
		message = "Bookmark name is empty."
		errorCode = 1
	} else if old.Name == bookmark.Name && old.URL == bookmark.URL &&
		string(old.Category) == bookmark.Category {
		message = "New bookmark is same as old bookmark."
	} else if err := db.QueryRow("SELECT id FROM bookmark WHERE bookmark = ? AND id != ? AND user_id = ?",
		bookmark.Name, id, userID).Scan(&exist); err == nil {
		message = fmt.Sprintf("Bookmark name %s is already existed.", bookmark.Name)
		errorCode = 1
	} else if err := db.QueryRow("SELECT id FROM bookmark WHERE url = ? AND id != ? AND user_id = ?",
		bookmark.URL, id, userID).Scan(&exist); err == nil {
		message = fmt.Sprintf("Bookmark url %s is already existed.", bookmark.URL)
		errorCode = 2
	} else if categoryID == -1 {
		message = "Category name exceeded length limit."
		errorCode = 3
	} else {
		if _, err := db.Exec("UPDATE bookmark SET bookmark = ?, url = ?, category_id = ? WHERE id = ? AND user_id = ?",
			bookmark.Name, bookmark.URL, categoryID, id, userID); err != nil {
			log.Println("Failed to edit bookmark:", err)
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
		log.Println("Failed to connect to database:", err)
		c.String(503, "")
		return
	}
	defer db.Close()

	session := sessions.Default(c)
	userID := session.Get("user_id")

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println("Failed to get id param:", err)
		c.String(400, "")
		return
	}

	if _, err := db.Exec("DELETE FROM bookmark WHERE id = ? and user_id = ?", id, userID); err != nil {
		log.Println("Failed to delete bookmark:", err)
		c.String(500, "")
		return
	}
	c.JSON(200, gin.H{"status": 1})
}

func reorder(c *gin.Context) {
	db, err := getDB()
	if err != nil {
		log.Println("Failed to connect to database:", err)
		c.String(503, "")
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")

	var reorder struct{ Old, New int }
	if err := c.BindJSON(&reorder); err != nil {
		c.String(400, "")
		return
	}

	var oldSeq, newSeq int
	if err := db.QueryRow("SELECT seq FROM seq WHERE bookmark_id = ? AND user_id = ?",
		reorder.Old, userID).Scan(&oldSeq); err != nil {
		log.Println("Failed to scan old seq:", err)
		c.String(500, "")
		return
	}
	if err := db.QueryRow("SELECT seq FROM seq WHERE bookmark_id = ? AND user_id = ?",
		reorder.New, userID).Scan(&newSeq); err != nil {
		log.Println("Failed to scan new seq:", err)
		c.String(500, "")
		return
	}

	if oldSeq > newSeq {
		_, err = db.Exec("UPDATE seq SET seq = seq+1 WHERE seq >= ? AND seq < ? AND user_id = ?",
			newSeq, oldSeq, userID)
	} else {
		_, err = db.Exec("UPDATE seq SET seq = seq-1 WHERE seq > ? AND seq <= ? AND user_id = ?",
			oldSeq, newSeq, userID)
	}
	if err != nil {
		log.Println("Failed to update other seq:", err)
		c.String(500, "")
		return
	}
	if _, err := db.Exec("UPDATE seq SET seq = ? WHERE bookmark_id = ? AND user_id = ?",
		newSeq, reorder.Old, userID); err != nil {
		log.Println("Failed to update seq:", err)
		c.String(500, "")
		return
	}
	c.String(200, "1")
}
