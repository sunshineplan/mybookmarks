package main

import (
	"log"
	"strings"

	"github.com/sunshineplan/database/mongodb/api"
)

func addUser(username string) {
	log.Print("Start!")
	if err := initDB(); err != nil {
		log.Fatalln("Failed to initialize database:", err)
	}

	username = strings.TrimSpace(strings.ToLower(username))

	insertedID, err := accountClient.InsertOne(api.M{
		"username": username,
		"password": "123456",
		"uid":      username,
	})
	if err != nil {
		log.Fatal(err)
	}

	if _, err := bookmarkClient.InsertOne(api.M{
		"bookmark": "Google",
		"url":      "https://www.google.com",
		"user":     insertedID,
		"seq":      1,
	}); err != nil {
		log.Fatal(err)
	}
	log.Print("Done!")
}

func deleteUser(username string) {
	log.Print("Start!")
	if err := initDB(); err != nil {
		log.Fatalln("Failed to initialize database:", err)
	}

	username = strings.TrimSpace(strings.ToLower(username))

	deletedCount, err := accountClient.DeleteOne(api.M{"username": username})
	if err != nil {
		log.Fatalln("Failed to delete user:", err)
	} else if deletedCount == 0 {
		log.Fatalf("User %s does not exist.", username)
	}
	log.Print("Done!")
}

func reorderBookmark(userID interface{}, orig, dest string) error {
	var origBookmark, destBookmark bookmark

	c := make(chan error, 1)
	go func() {
		c <- bookmarkClient.FindOne(api.M{"_id": api.ObjectID(orig)}, nil, &origBookmark)
	}()
	if err := bookmarkClient.FindOne(api.M{"_id": api.ObjectID(dest)}, nil, &destBookmark); err != nil {
		return err
	}
	if err := <-c; err != nil {
		return err
	}

	var filter, update api.M
	if origBookmark.Seq > destBookmark.Seq {
		filter = api.M{"user": userID, "seq": api.M{"$gte": destBookmark.Seq, "$lt": origBookmark.Seq}}
		update = api.M{"$inc": api.M{"seq": 1}}
	} else {
		filter = api.M{"user": userID, "seq": api.M{"$gt": origBookmark.Seq, "$lte": destBookmark.Seq}}
		update = api.M{"$inc": api.M{"seq": -1}}
	}

	if _, err := bookmarkClient.UpdateMany(filter, update, nil); err != nil {
		log.Println("Failed to reorder bookmark:", err)
		return err
	}

	if _, err := bookmarkClient.UpdateOne(
		api.M{"_id": api.ObjectID(orig)},
		api.M{"$set": api.M{"seq": destBookmark.Seq}},
		nil,
	); err != nil {
		log.Println("Failed to reorder bookmark:", err)
		return err
	}

	return nil
}
