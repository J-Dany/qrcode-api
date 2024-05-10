package database

import (
	"context"
	"os"
	"qrcode/env"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var MongoDatabase string

func init() {
	env.LoadEnvs()

	mongodbUri := os.Getenv("MONGODB_URI")
	MongoDatabase = os.Getenv("MONGO_INITDB_DATABASE")

	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongodbUri))

	if err != nil {
		panic(err)
	}

	Client = client
}
