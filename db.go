package main

import (
	"github.com/sunshineplan/utils/database/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
)

var dbConfig mongodb.Config
var collAccount *mongo.Collection
var collBookmark *mongo.Collection

func initDB() (err error) {
	if err = meta.Get("mybookmarks_mongo", &dbConfig); err != nil {
		return
	}

	var client *mongo.Client
	client, err = dbConfig.Open()
	if err != nil {
		return
	}

	database := client.Database(dbConfig.Database)

	if !universal {
		collAccount = database.Collection("account")
	}
	collBookmark = database.Collection("bookmark")

	return
}
