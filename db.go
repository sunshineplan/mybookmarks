package main

import (
	"github.com/sunshineplan/database/mongodb/api"
	"github.com/sunshineplan/utils"
)

var accountClient api.Client
var bookmarkClient api.Client

func initDB() error {
	var mongo api.Client
	if err := utils.Retry(func() error {
		return meta.Get("mybookmarks_mongo", &mongo)
	}, 3, 20); err != nil {
		return err
	}

	accountClient, bookmarkClient = mongo, mongo
	accountClient.Collection = "account"
	bookmarkClient.Collection = "bookmark"

	return nil
}

func test() error {
	var mongo api.Client
	return meta.Get("mybookmarks_mongo", &mongo)
}
