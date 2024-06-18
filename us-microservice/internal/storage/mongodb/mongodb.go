package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"
	"urlSh/internal/storage"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Storage struct {
	client     *mongo.Client
	collection *mongo.Collection
}

type URLDocument struct {
	Alias string `bson:"alias"`
	URL   string `bson:"url"`
}

func New(uri, database, collection string) (*Storage, error) {
	const op = "storage.mongodb.New"

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("%s: create client: %w", op, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("%s: ping database: %w", op, err)
	}

	db := client.Database(database)
	coll := db.Collection(collection)

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "alias", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err = coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return nil, fmt.Errorf("%s: create index: %w", op, err)
	}

	return &Storage{client: client, collection: coll}, nil
}

func (s *Storage) SaveURL(ctx context.Context, urlToSave, alias string) (string, error) {
	const op = "storage.mongodb.SaveURL"

	doc := URLDocument{
		Alias: alias,
		URL:   urlToSave,
	}

	_, err := s.collection.InsertOne(ctx, doc)
	if err != nil {
		var mongoWriteException mongo.WriteException
		if errors.As(err, &mongoWriteException) {
			for _, writeError := range mongoWriteException.WriteErrors {
				if writeError.Code == 11000 {
					return "", fmt.Errorf("%s: %w", op, storage.ErrURLExists)
				}
			}
		}
		return "", fmt.Errorf("%s: insert document: %w", op, err)
	}

	return doc.Alias, nil
}

func (s *Storage) GetURL(ctx context.Context, alias string) (string, error) {
	const op = "storage.mongodb.GetURL"

	var doc URLDocument
	filter := bson.D{{Key: "alias", Value: alias}}

	err := s.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: find document: %w", op, err)
	}

	return doc.URL, nil
}
