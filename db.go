package main

import (
	"github.com/sunshineplan/database/mongodb"
	"github.com/sunshineplan/database/mongodb/api"
	"github.com/sunshineplan/utils"
)

var accountClient mongodb.Client
var bookmarkClient mongodb.Client

func initDB() (err error) {
	var mongo api.Client
	if err = utils.Retry(func() error {
		return meta.Get("mybookmarks_mongo", &mongo)
	}, 3, 20); err != nil {
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
