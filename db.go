package main

import (
	"github.com/sunshineplan/database/mongodb"
	"github.com/sunshineplan/utils"
	"go.mongodb.org/mongo-driver/mongo"
)

var dbConfig mongodb.Config
var collAccount *mongo.Collection
var collBookmark *mongo.Collection

func initDB() error {
	if err := utils.Retry(func() error {
		return meta.Get("mybookmarks_mongo", &dbConfig)
	}, 3, 20); err != nil {
		return err
	}

	client, err := dbConfig.Open()
	if err != nil {
		return err
	}

	database := client.Database(dbConfig.Database)

	collAccount = database.Collection("account")
	collBookmark = database.Collection("bookmark")

	return nil
}

func test() (err error) {
	if err = meta.Get("mybookmarks_mongo", &dbConfig); err != nil {
		return
	}

	_, err = dbConfig.Open()
	return
}
