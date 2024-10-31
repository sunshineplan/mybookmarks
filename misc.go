package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/sunshineplan/database/mongodb"
)

func addUser(username string) error {
	svc.Print("Start!")
	if err := initDB(); err != nil {
		return err
	}

	username = strings.TrimSpace(strings.ToLower(username))

	insertedID, err := accountClient.InsertOne(
		struct {
			Username string `json:"username" bson:"username"`
			Password string `json:"password" bson:"password"`
			Uid      string `json:"uid" bson:"uid"`
		}{username, "123456", username},
	)
	if err != nil {
		return err
	}

	if _, err := bookmarkClient.InsertOne(
		struct {
			Bookmark string `json:"bookmark" bson:"bookmark"`
			URL      string `json:"url" bson:"url"`
			User     string `json:"user" bson:"user"`
			Seq      int    `json:"seq" bson:"seq"`
			Created  any    `json:"created" bson:"created"`
		}{"Google", "https://www.google.com", insertedID.(mongodb.ObjectID).Hex(), 1, bookmarkClient.Date(time.Now())},
	); err != nil {
		return err
	}
	svc.Print("Done!")
	return nil
}

func deleteUser(username string) error {
	svc.Print("Start!")
	if err := initDB(); err != nil {
		return err
	}

	username = strings.TrimSpace(strings.ToLower(username))

	deletedCount, err := accountClient.DeleteOne(mongodb.M{"username": username})
	if err != nil {
		return err
	} else if deletedCount == 0 {
		return fmt.Errorf("user %s does not exist", username)
	}
	svc.Print("Done!")
	return nil
}

func checkExist(filter any) (ok bool, err error) {
	var exist any
	err = bookmarkClient.FindOne(filter, nil, &exist)
	ok = err == nil
	if err == mongodb.ErrNoDocuments {
		err = nil
	}
	return
}

func reorderBookmark(userID string, orig, dest mongodb.ObjectID) error {
	var origBookmark, destBookmark bookmark

	c := make(chan error, 1)
	go func() {
		c <- bookmarkClient.FindOne(mongodb.M{"_id": orig}, nil, &origBookmark)
	}()
	if err := bookmarkClient.FindOne(mongodb.M{"_id": dest}, nil, &destBookmark); err != nil {
		return err
	}
	if err := <-c; err != nil {
		return err
	}

	var filter, update mongodb.M
	if origBookmark.Seq > destBookmark.Seq {
		filter = mongodb.M{"user": userID, "seq": mongodb.M{"$gte": destBookmark.Seq, "$lt": origBookmark.Seq}}
		update = mongodb.M{"$inc": mongodb.M{"seq": 1}}
	} else {
		filter = mongodb.M{"user": userID, "seq": mongodb.M{"$gt": origBookmark.Seq, "$lte": destBookmark.Seq}}
		update = mongodb.M{"$inc": mongodb.M{"seq": -1}}
	}

	if _, err := bookmarkClient.UpdateMany(filter, update, nil); err != nil {
		return err
	}

	if _, err := bookmarkClient.UpdateOne(
		mongodb.M{"_id": orig},
		mongodb.M{"$set": mongodb.M{"seq": destBookmark.Seq}},
		nil,
	); err != nil {
		return err
	}

	return nil
}
