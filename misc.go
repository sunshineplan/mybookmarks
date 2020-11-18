package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sunshineplan/utils/mail"
)

func addUser(username string) {
	log.Println("Start!")
	db, err := getDB()
	if err != nil {
		log.Fatalln("Failed to connect to database:", err)
	}
	defer db.Close()
	if _, err := db.Exec("INSERT INTO user(username) VALUES (?)", strings.ToLower(username)); err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			log.Fatalf("Username %s already exists.", strings.ToLower(username))
		} else {
			log.Fatalln("Failed to add user:", err)
		}
	}
	log.Println("Done!")
}

func deleteUser(username string) {
	log.Println("Start!")
	db, err := getDB()
	if err != nil {
		log.Fatalln("Failed to connect to database:", err)
	}
	defer db.Close()
	res, err := db.Exec("DELETE FROM user WHERE username=?", strings.ToLower(username))
	if err != nil {
		log.Fatalln("Failed to delete user:", err)
	}
	if n, err := res.RowsAffected(); err != nil {
		log.Fatalln("Failed to get affected rows:", err)
	} else if n == 0 {
		log.Fatalf("User %s does not exist.", strings.ToLower(username))
	}
	log.Println("Done!")
}

func backup() {
	log.Println("Start!")
	var config struct {
		SMTPServer     string
		SMTPServerPort int
		From, Password string
		To             []string
	}
	if err := meta.Get("mybookmarks_backup", &config); err != nil {
		log.Fatalln("Failed to get mybookmarks_backup metadata:", err)
	}
	dialer := &mail.Dialer{
		Host:     config.SMTPServer,
		Port:     config.SMTPServerPort,
		Account:  config.From,
		Password: config.Password,
	}

	file := dump()
	defer os.Remove(file)
	if err := dialer.Send(
		&mail.Message{
			To:          config.To,
			Subject:     fmt.Sprintf("My Bookmarks Backup-%s", time.Now().Format("20060102")),
			Attachments: []*mail.Attachment{{Path: file, Filename: "database"}},
		},
	); err != nil {
		log.Fatalln("Failed to send mail:", err)
	}
	log.Println("Done!")
}

func restore(file string) {
	log.Println("Start!")
	if file == "" {
		file = joinPath(dir(self), "scripts/schema.sql")
	} else {
		if _, err := os.Stat(file); err != nil {
			log.Fatalln("File not found:", err)
		}
	}
	dropAll := joinPath(dir(self), "scripts/drop_all.sql")
	execScript(dropAll)
	execScript(file)
	log.Println("Done!")
}
