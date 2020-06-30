package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sunshineplan/metadata"
	"github.com/sunshineplan/utils/mail"
)

func addUser(username string) {
	log.Println("Start!")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	_, err = db.Exec("INSERT INTO user(username) VALUES (?)", strings.ToLower(username))
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			log.Fatalf("Username %s already exists.", strings.ToLower(username))
		} else {
			log.Fatalf("Failed to add user: %v", err)
		}
	}
	log.Println("Done!")
}

func deleteUser(username string) {
	log.Println("Start!")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	res, err := db.Exec("DELETE FROM user WHERE username=?", strings.ToLower(username))
	if err != nil {
		log.Fatalf("Failed to delete user: %v", err)
	}
	if n, err := res.RowsAffected(); err != nil {
		log.Fatalf("Failed to get affected rows: %v", err)
	} else if n == 0 {
		log.Fatalf("User %s does not exist.", strings.ToLower(username))
	}
	log.Println("Done!")
}

func backup() {
	log.Println("Start!")
	m, err := metadata.Get("mybookmarks_backup", &metadataConfig)
	if err != nil {
		log.Fatalf("Failed to get mybookmarks_backup metadata: %v", err)
	}
	var mailSetting mail.Setting
	err = json.Unmarshal(m, &mailSetting)
	if err != nil {
		log.Fatalf("Failed to unmarshal json: %v", err)
	}

	file := Dump()
	defer os.Remove(file)
	err = mail.SendMail(
		&mailSetting,
		fmt.Sprintf("My Bookmarks Backup-%s", time.Now().Format("20060102")),
		"",
		&mail.Attachment{FilePath: file, Filename: "database"},
	)
	if err != nil {
		log.Fatalf("Failed to send mail: %v", err)
	}
	log.Println("Done!")
}
