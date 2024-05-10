package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id"`
	Username string             `bson:"username"`
}

func InsertUserIfNotExists(username string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user User
	userDb := Client.Database(MongoDatabase).Collection("users").FindOne(context.Background(), map[string]interface{}{"username": username})
	var decodeErr = userDb.Decode(&user)

	if decodeErr == nil {
		return &user, nil
	}

	result, err := Client.Database(MongoDatabase).Collection("users").InsertOne(ctx, map[string]interface{}{"username": username})

	if err != nil {
		return nil, err
	}

	decodeErr = Client.Database(MongoDatabase).Collection("users").FindOne(ctx, bson.D{{Key: "_id", Value: result.InsertedID}}).Decode(&user)

	if decodeErr != nil {
		return nil, decodeErr
	}

	return &user, nil
}
