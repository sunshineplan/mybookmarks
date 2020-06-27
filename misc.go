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
		log.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec("INSERT INTO user(username) VALUES (?)", strings.ToLower(username))
	if err != nil {
		log.Fatalf("[ERROR]Username %s already exists.\n", strings.ToLower(username))
	}
	log.Println("Done!")
}

func deleteUser(username string) {
	log.Println("Start!")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	res, _ := db.Exec("DELETE FROM user WHERE username=?", strings.ToLower(username))
	if n, _ := res.RowsAffected(); n != 0 {
		log.Println("Done.")
	} else {
		log.Fatalf("[ERROR]User %s does not exist.\n", strings.ToLower(username))
	}
	log.Println("Done!")
}

func backup() {
	log.Println("Start!")
	m, err := metadata.Get("mybookmarks_backup", &metadataConfig)
	if err != nil {
		log.Fatal(err)
	}
	var mailSetting mail.Setting
	err = json.Unmarshal(m, &mailSetting)
	if err != nil {
		log.Fatalln(err)
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
		log.Fatal(err)
	}
	log.Println("Done!")
}
