package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type bookmark struct {
	ID       int    `json:"id"`
	Name     string `json:"bookmark"`
	URL      string `json:"url"`
	Category string `json:"category"`
}

func checkBookmark(bookmarkID, userID interface{}) bool {
	var exist string
	if err := db.QueryRow("SELECT bookmark FROM bookmarks WHERE bookmark_id = ? AND user_id = ?",
		bookmarkID, userID).Scan(&exist); err == nil {
		return true
	}
	return false
}

func getBookmark(userID interface{}) (bookmarks []bookmark, total int, err error) {
	ec := make(chan error, 1)
	go func() {
		ec <- db.QueryRow("SELECT count(bookmark) FROM bookmark WHERE user_id = ?", userID).Scan(&total)
	}()

	bookmarks = []bookmark{}
	if userID == nil {
		return
	}

	var rows *sql.Rows
	rows, err = db.Query("SELECT bookmark_id, bookmark, url, category FROM bookmarks WHERE user_id = ? LIMIT 50", userID)
	if err != nil {
		log.Println("Failed to get bookmarks:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var bookmark bookmark
		var categoryByte []byte
		if err = rows.Scan(&bookmark.ID, &bookmark.Name, &bookmark.URL, &categoryByte); err != nil {
			log.Println("Failed to scan bookmark:", err)
			return
		}
		bookmark.Category = string(categoryByte)
		bookmarks = append(bookmarks, bookmark)
	}
	err = <-ec

	return
}

func moreBookmark(c *gin.Context) {
	var r struct{ Start int }
	if err := c.BindJSON(&r); err != nil {
		c.String(400, "")
		return
	}

	rows, err := db.Query("SELECT bookmark_id, bookmark, url, category FROM bookmarks WHERE user_id = ? LIMIT ?, 50",
		sessions.Default(c).Get("userID"), r.Start)
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
			log.Println("Failed to scan bookmark:", err)
			c.String(500, "")
			return
		}
		bookmark.Category = string(categoryByte)
		bookmarks = append(bookmarks, bookmark)
	}
	c.JSON(200, bookmarks)
}

func addBookmark(c *gin.Context) {
	var bookmark bookmark
	if err := c.BindJSON(&bookmark); err != nil {
		c.String(400, "")
		return
	}

	userID := sessions.Default(c).Get("userID")
	bc := make(chan error, 3)
	var categoryID int
	var exist1, exist2 string
	go func() {
		var err error
		categoryID, err = getCategoryID(bookmark.Category, userID.(int))
		bc <- err
	}()
	go func() {
		db.QueryRow("SELECT id FROM bookmark WHERE bookmark = ? AND user_id = ?",
			bookmark.Name, userID).Scan(&exist1)
		bc <- nil
	}()
	go func() {
		db.QueryRow("SELECT id FROM bookmark WHERE url = ? AND user_id = ?",
			bookmark.URL, userID).Scan(&exist2)
		bc <- nil
	}()
	for i := 0; i < 3; i++ {
		if err := <-bc; err != nil {
			log.Println("Failed to get category id:", err)
			c.String(500, "")
			return
		}
	}

	var message string
	var errorCode int
	switch {
	case bookmark.Name == "":
		message = "Bookmark name is empty."
		errorCode = 1
	case exist1 != "":
		message = fmt.Sprintf("Bookmark name %s is already existed.", bookmark.Name)
		errorCode = 1
	case exist2 != "":
		message = fmt.Sprintf("Bookmark url %s is already existed.", bookmark.URL)
		errorCode = 2
	case categoryID == -1:
		message = "Category name exceeded length limit."
		errorCode = 3
	default:
		res, err := db.Exec("INSERT INTO bookmark (bookmark, url, user_id, category_id) VALUES (?, ?, ?, ?)",
			bookmark.Name, bookmark.URL, userID, categoryID)
		if err != nil {
			log.Println("Failed to add bookmark:", err)
			c.String(500, "")
			return
		}
		id, err := res.LastInsertId()
		if err != nil {
			log.Println("Failed to get last insert id:", err)
			c.String(500, "")
			return
		}
		c.JSON(200, gin.H{"status": 1, "id": id, "cid": categoryID})
		return
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}

func editBookmark(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println("Failed to get id param:", err)
		c.String(400, "")
		return
	}

	var new bookmark
	if err := c.BindJSON(&new); err != nil {
		c.String(400, "")
		return
	}

	userID := sessions.Default(c).Get("userID")
	bc := make(chan error, 4)
	var old bookmark
	var categoryID int
	var exist1, exist2 string
	go func() {
		var oldCategory []byte
		err := db.QueryRow("SELECT bookmark, url, category FROM bookmarks WHERE bookmark_id = ? AND user_id = ?",
			id, userID).Scan(&old.Name, &old.URL, &oldCategory)
		old.Category = string(oldCategory)
		bc <- err
	}()
	go func() {
		var err error
		categoryID, err = getCategoryID(new.Category, userID.(int))
		bc <- err
	}()
	go func() {
		db.QueryRow("SELECT id FROM bookmark WHERE bookmark = ? AND id != ? AND user_id = ?",
			new.Name, id, userID).Scan(&exist1)
		bc <- nil
	}()
	go func() {
		db.QueryRow("SELECT id FROM bookmark WHERE url = ? AND id != ? AND user_id = ?",
			new.URL, id, userID).Scan(&exist2)
		bc <- nil
	}()
	for i := 0; i < 4; i++ {
		if err := <-bc; err != nil {
			log.Println(err)
			c.String(500, "")
			return
		}
	}

	var message string
	var errorCode int
	switch {
	case new.Name == "":
		message = "Bookmark name is empty."
		errorCode = 1
	case old == new:
		message = "New bookmark is same as old bookmark."
	case exist1 != "":
		message = fmt.Sprintf("Bookmark name %s is already existed.", new.Name)
		errorCode = 1
	case exist2 != "":
		message = fmt.Sprintf("Bookmark url %s is already existed.", new.URL)
		errorCode = 2
	case categoryID == -1:
		message = "Category name exceeded length limit."
		errorCode = 3
	default:
		if _, err := db.Exec("UPDATE bookmark SET bookmark = ?, url = ?, category_id = ? WHERE id = ? AND user_id = ?",
			new.Name, new.URL, categoryID, id, userID); err != nil {
			log.Println("Failed to edit bookmark:", err)
			c.String(500, "")
			return
		}
		c.JSON(200, gin.H{"status": 1, "cid": categoryID})
		return
	}
	c.JSON(200, gin.H{"status": 0, "message": message, "error": errorCode})
}

func deleteBookmark(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println("Failed to get id param:", err)
		c.String(400, "")
		return
	}

	if checkBookmark(id, sessions.Default(c).Get("userID")) {
		if _, err := db.Exec("DELETE FROM bookmark WHERE id = ? and user_id = ?",
			id, sessions.Default(c).Get("userID")); err != nil {
			log.Println("Failed to delete bookmark:", err)
			c.String(500, "")
			return
		}
		c.JSON(200, gin.H{"status": 1})
		return
	}
	c.String(403, "")
}

func reorder(c *gin.Context) {
	var reorder struct{ Old, New int }
	if err := c.BindJSON(&reorder); err != nil {
		c.String(400, "")
		return
	}

	userID := sessions.Default(c).Get("userID")
	if !checkBookmark(reorder.Old, userID) || !checkBookmark(reorder.New, userID) {
		c.String(403, "")
		return
	}

	if _, err := db.Exec("CALL reorder(?, ?, ?)",
		userID, reorder.New, reorder.Old); err != nil {
		log.Println("Failed to reorder bookmark:", err)
		c.String(500, "")
		return
	}
	c.String(200, "1")
}
