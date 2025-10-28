package main

import (
	"time"

	"github.com/sunshineplan/database/mongodb"
	"github.com/sunshineplan/database/mongodb/driver"
	"github.com/sunshineplan/utils/retry"
)

var accountClient mongodb.Client
var bookmarkClient mongodb.Client

func initDB() (err error) {
	var mongo driver.Client
	if err = retry.Do(func() error {
		return meta.Get("mybookmarks_mongo", &mongo)
	}, 3, 20*time.Second); err != nil {
		return err
	}

	account, bookmark := mongo, mongo
	account.Collection = "account"
	bookmark.Collection = "bookmark"
	accountClient, bookmarkClient = &account, &bookmark

	if err = accountClient.Connect(); err != nil {
		return
	}
	return bookmarkClient.Connect()
}

func test() error {
	return initDB()
}
