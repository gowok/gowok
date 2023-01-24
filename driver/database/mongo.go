package database

import (
	"context"

	"github.com/gowok/gowok/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoClient interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Database(name string, opts ...*options.DatabaseOptions) *mongo.Database
	Ping(ctx context.Context, rp *readpref.ReadPref) error
}

type MongoDatabase interface {
	CreateCollection(ctx context.Context, name string, opts ...*options.CreateCollectionOptions) error
	ListCollectionNames(ctx context.Context, filter any, opts ...*options.ListCollectionsOptions) ([]string, error)
	Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection
	Drop(ctx context.Context) error
}

type MongoInserter interface {
	InsertMany(ctx context.Context, documents []any, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error)
	InsertOne(ctx context.Context, document any, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
}

type MongoAggregator interface {
	Aggregate(ctx context.Context, pipeline any, opts ...*options.AggregateOptions) (*mongo.Cursor, error)
}

type MongoFinder interface {
	Find(ctx context.Context, filter any, opts ...*options.FindOptions) (*mongo.Cursor, error)
	FindOne(ctx context.Context, filter any, opts ...*options.FindOneOptions) *mongo.SingleResult
	Distinct(ctx context.Context, fieldName string, filter any, opts ...*options.DistinctOptions) ([]any, error)
	CountDocuments(ctx context.Context, filter any, opts ...*options.CountOptions) (int64, error)
}

type MongoUpdater interface {
	UpdateMany(ctx context.Context, filter any, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	UpdateOne(ctx context.Context, filter any, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	UpdateByID(ctx context.Context, id any, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
}

type MongoDeleter interface {
	DeleteMany(ctx context.Context, filter any, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
	DeleteOne(ctx context.Context, filter any, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
}

func NewMongo(conf config.Database) (*mongo.Client, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(conf.DSN))
	if err != nil {
		return nil, err
	}

	return client, err
}
