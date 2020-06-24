package main

import (
	"database/sql"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func addUser(username string) {
	db, err := sql.Open("mysql", "user:password@/dbname")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec("INSERT INTO user(username) VALUES (?)", strings.ToLower(username))
	if err != nil {
		log.Fatalf("[ERROR]Username %s already exists.\n", strings.ToLower(username))
	}
	log.Println("Done.")
}

func deleteUser(username string) {
	db, err := sql.Open("mysql", "user:password@/dbname")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	res, _ := db.Exec("DELETE FROM user WHERE username=?", strings.ToLower(username))
	if n, _ := res.RowsAffected(); n != 0 {
		log.Println("Done.")
	} else {
		log.Fatalf("[ERROR]User {username.lower()} does not exist.\n", strings.ToLower(username))
	}
}

func restore(filePath string) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("mysql", "user:password@/dbname")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	dropAll, err := ioutil.ReadFile(filepath.Join(filepath.Dir(self), "drop_all.sql"))
	if err != nil {
		log.Fatal(err)
	}
	tx, _ := db.Begin()
	_, err = tx.Exec(string(dropAll))
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Fatal(rollbackErr)
		}
		log.Fatal(err)
	}
	_, err = tx.Exec(string(file))
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Fatal(rollbackErr)
		}
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
	log.Println("Done.")
}

//def backup():
//    try:
//        msg = EmailMessage()
//        msg['Subject'] = f'My Bookmarks Backup-{datetime.now():%Y%m%d}'
//        msg['From'] = BACKUP['sender']
//        msg['To'] = BACKUP['subscriber']
//        mem = StringIO()
//        db = sqlite3.connect(app.config['DATABASE'])
//        mem.write('\n'.join(db.iterdump()))
//        db.close()
//        msg.add_attachment(mem.getvalue(), filename='database')
//        mem.close()
//        with SMTP(BACKUP['smtp_server'], BACKUP['smtp_server_port']) as s:
//            s.starttls()
//            s.login(BACKUP['sender'], BACKUP['password'])
//            s.send_message(msg)
//        click.echo('Done.')
//    except:
//        click.echo('Failed. Please check mail setting.')
//
//
//
