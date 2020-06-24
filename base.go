package main

import "github.com/gin-gonic/gin"

func modeAction(c *gin.Context) {
	action := c.Param("action")
	switch mode := c.Param("mode"); mode {
	case "bookmark":
		switch action {
		case "add":
			addBookmark(c)
			return
		}
	case "category":
		switch action {
		case "add":
			addCategory(c)
			return
		case "get":
			getCategory(c)
			return
		}
	}
	c.String(400, "")
}

func doModeAction(c *gin.Context) {
	mode := c.Param("mode")
	if action := c.Param("action"); action == "add" {
		switch mode {
		case "bookmark":
			doAddBookmark(c)
			return
		case "category":
			doAddCategory(c)
			return
		}
	}
	c.String(400, "")
}

func modeActionID(c *gin.Context) {
	mode := c.Param("mode")
	if action := c.Param("action"); action == "edit" {
		switch mode {
		case "bookmark":
			editBookmark(c)
			return
		case "category":
			editCategory(c)
			return
		}
	}
	c.String(400, "")
}

func doModeActionID(c *gin.Context) {
	action := c.Param("action")
	switch mode := c.Param("mode"); mode {
	case "bookmark":
		switch action {
		case "edit":
			doEditBookmark(c)
			return
		case "delete":
			doDeleteBookmark(c)
			return
		}
	case "category":
		switch action {
		case "edit":
			doEditCategory(c)
			return
		case "delete":
			doDeleteCategory(c)
			return
		}
	}
	c.String(400, "")
}
