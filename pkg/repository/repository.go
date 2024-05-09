package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository struct {
	db *mongo.Database
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) GetCollection(collectionName string) *mongo.Collection {
	return r.db.Collection(collectionName)
}

func (r *Repository) GetDatabase() *mongo.Database {
	return r.db
}

func (r *Repository) Get(collectionName string, filter interface{}, options ...*options.FindOneOptions) *mongo.SingleResult {
	return r.GetCollection(collectionName).FindOne(context.Background(), filter, options...)
}

func (r *Repository) GetAll(collectionName string, filter interface{}, options ...*options.FindOptions) (*mongo.Cursor, error) {
	return r.GetCollection(collectionName).Find(context.Background(), filter, options...)
}

func (r *Repository) Insert(collectionName string, document interface{}, options ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return r.GetCollection(collectionName).InsertOne(context.Background(), document, options...)
}

func (r *Repository) InsertAll(collectionName string, documents []interface{}, options ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	return r.GetCollection(collectionName).InsertMany(context.Background(), documents, options...)
}

func (r *Repository) Update(collectionName string, filter interface{}, update interface{}, options ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return r.GetCollection(collectionName).UpdateOne(context.Background(), filter, update, options...)
}

func (r *Repository) UpdateAll(collectionName string, filter interface{}, update interface{}, options ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return r.GetCollection(collectionName).UpdateMany(context.Background(), filter, update, options...)
}
