package mongodb

import (
	"context"
	"d2api/config"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

var Instance *MongoDB

func Init(mongoConfig *config.MongoConfig) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(mongoConfig.URL).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	if err := client.Ping(context.Background(), nil); err != nil {
		log.Println("Failed to ping MongoDB, error", err.Error())
		panic(err)
	}

	log.Println("Successfully connected to MongoDB!")

	Instance = &MongoDB{
		Client:   client,
		Database: client.Database(mongoConfig.Database),
	}
}

func GetCollection(collection string) *mongo.Collection {
	return Instance.Database.Collection(collection)
}
