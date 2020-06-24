package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func getBookmark(c *gin.Context) {
	db, err := sql.Open("mysql", "user:password@/dbname")
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()
	session := sessions.Default(c)
	userID := session.Get("user_id")

	type Bookmark struct {
		id       int
		name     string
		url      string
		category string
	}
	var category map[string]interface{}
	var bookmarks []Bookmark
	switch categoryID := c.Query("category"); categoryID {
	case "":
		category = map[string]interface{}{"id": -1, "name": "All Bookmarks"}
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
			var bookmark Bookmark
			if err := rows.Scan(&bookmark.id, &bookmark.name, &bookmark.url, &bookmark.category); err != nil {
				log.Println(err)
				return
			}
			bookmarks = append(bookmarks, bookmark)
		}
	case "0":
		category = map[string]interface{}{"id": 0, "name": "Uncategorized"}
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
			var bookmark Bookmark
			if err := rows.Scan(&bookmark.id, &bookmark.name, &bookmark.url); err != nil {
				log.Println(err)
				return
			}
			bookmarks = append(bookmarks, bookmark)
		}
	default:
		category = map[string]interface{}{"id": categoryID}
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
		for rows.Next() {
			var bookmark Bookmark
			if err := rows.Scan(&bookmark.id, &bookmark.name, &bookmark.url); err != nil {
				log.Println(err)
				return
			}
			bookmarks = append(bookmarks, bookmark)
		}
		defer rows.Close()
		for _, b := range bookmarks {
			b.category = category["name"].(string)
		}
		c.HTML(200, "bookmark/index.html", gin.H{"category": category, "bookmarks": bookmarks})
	}
}

//
//
//def get_category_id(category, user_id):
//    if category:
//        db = get_db()
//        category_id = db.execute(
//            "SELECT id FROM category WHERE category = ? AND user_id = ?", (category, user_id)).fetchone()
//        if len(category.encode("utf-8")) > 15:
//            return None
//        elif category_id:
//            return category_id["id"]
//        else:
//            db.execute(
//                "INSERT INTO category (category, user_id) VALUES (?, ?)", (category, user_id))
//            return db.execute("SELECT last_insert_rowid() id").fetchone()["id"]
//    else:
//        return 0
//
//
//@bp.route("/bookmark/add", methods=("GET", "POST"))
//@login_required
//def add_bookmark():
//    """Create a new bookmark for the current user."""
//    category_id = request.args.get("category_id")
//    db = get_db()
//    if category_id:
//        category = db.execute("SELECT category FROM category WHERE id = ? AND user_id = ?",
//                              (category_id, g.user["id"])).fetchone()["category"]
//    else:
//        category = ""
//    categories = db.execute(
//        "SELECT category FROM category WHERE user_id = ? ORDER BY category", (g.user["id"],)).fetchall()
//    if request.method == "POST":
//        category = request.form.get("category").strip()
//        bookmark = request.form.get("bookmark").strip()
//        url = request.form.get("url").strip()
//        category_id = get_category_id(category, g.user["id"])
//        error = 0
//        if bookmark == "":
//            message = f"Bookmark name is empty."
//            error = 1
//        elif db.execute("SELECT id FROM bookmark WHERE bookmark = ? AND user_id = ?", (bookmark, g.user["id"])).fetchone() is not None:
//            message = f"Bookmark name {bookmark} is already existed."
//            error = 1
//        elif db.execute("SELECT id FROM bookmark WHERE url = ? AND user_id = ?", (url, g.user["id"])).fetchone() is not None:
//            message = f"Bookmark url {url} is already existed."
//            error = 2
//        elif category_id is None:
//            message = "Category name exceeded length limit."
//            error = 3
//        else:
//            db.execute("INSERT INTO bookmark (bookmark, url, user_id, category_id)"
//                       " VALUES (?, ?, ?, ?)", (bookmark, url, g.user["id"], category_id))
//            db.commit()
//            return jsonify({"status": 1})
//        return jsonify({"status": 0, "message": message, "error": error})
//    return render_template("bookmark/bookmark.html", id=0, bookmark={"category": category}, categories=categories)
//
//
//@bp.route("/bookmark/edit/<int:id>", methods=("GET", "POST"))
//@login_required
//def edit_bookmark(id):
//    """Edit a bookmark for the current user."""
//    db = get_db()
//    bookmark = db.execute("SELECT bookmark, url, category FROM bookmark"
//                          " LEFT JOIN category ON category_id = category.id"
//                          " WHERE bookmark.id = ? AND bookmark.user_id = ?",
//                          (id, g.user["id"])).fetchone()
//    if not bookmark:
//        abort(403)
//    else:
//        if not bookmark["category"]:
//            bookmark["category"] = ""
//    categories = db.execute(
//        "SELECT category FROM category WHERE user_id = ? ORDER BY category", (g.user["id"],)).fetchall()
//    if request.method == "POST":
//        old = (bookmark["bookmark"], bookmark["url"], bookmark["category"])
//        bookmark = request.form.get("bookmark").strip()
//        url = request.form.get("url").strip()
//        category = request.form.get("category").strip()
//        category_id = get_category_id(category, g.user["id"])
//        error = 0
//        if bookmark == "":
//            message = f"Bookmark name is empty."
//            error = 1
//        elif old == (bookmark, url, category):
//            message = "New bookmark is same as old bookmark."
//        elif db.execute("SELECT id FROM bookmark WHERE bookmark = ? AND id != ? AND user_id = ?", (bookmark, id, g.user["id"])).fetchone() is not None:
//            message = f"Bookmark name {bookmark} is already existed."
//            error = 1
//        elif db.execute("SELECT id FROM bookmark WHERE url = ? AND id != ? AND user_id = ?", (url, id, g.user["id"])).fetchone() is not None:
//            message = f"Bookmark url {url} is already existed."
//            error = 2
//        elif category_id is None:
//            message = "Category name exceeded length limit."
//            error = 3
//        else:
//            db.execute("UPDATE bookmark SET bookmark = ?, url = ?, category_id = ?"
//                       " WHERE id = ? AND user_id = ?", (bookmark, url, category_id, id, g.user["id"]))
//            db.commit()
//            return jsonify({"status": 1})
//        return jsonify({"status": 0, "message": message, "error": error})
//    return render_template("bookmark/bookmark.html", id=id, bookmark=bookmark, categories=categories)
//
//
//@bp.route("/bookmark/delete/<int:id>", methods=("POST",))
//@login_required
//def delete_bookmark(id):
//    """Edit a bookmark for the current user."""
//    db = get_db()
//    db.execute("DELETE FROM bookmark WHERE id = ? and user_id = ?",
//               (id, g.user["id"]))
//    db.commit()
//    return jsonify({"status": 1})
//
//

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
