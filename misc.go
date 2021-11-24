package main

import (
	"log"
	"strings"
	"time"

	"github.com/sunshineplan/database/mongodb/api"
)

func addUser(username string) {
	log.Print("Start!")
	if err := initDB(); err != nil {
		log.Fatalln("Failed to initialize database:", err)
	}

	username = strings.TrimSpace(strings.ToLower(username))

	insertedID, err := accountClient.InsertOne(
		struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Uid      string `json:"uid"`
		}{username, "123456", username},
	)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := bookmarkClient.InsertOne(
		struct {
			Bookmark string      `json:"bookmark"`
			URL      string      `json:"url"`
			User     string      `json:"user"`
			Seq      int         `json:"seq"`
			Created  interface{} `json:"created"`
		}{"Google", "https://www.google.com", insertedID, 1, api.Date(time.Now())},
	); err != nil {
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
