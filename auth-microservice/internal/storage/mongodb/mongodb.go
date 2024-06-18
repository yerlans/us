package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"auth/internal/domain/models"
	"auth/internal/storage"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Storage struct {
	client     *mongo.Client
	collection *mongo.Collection
}

type UserDocument struct {
	Email    string `bson:"email"`
	Password string `bson:"password"`
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
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err = coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return nil, fmt.Errorf("%s: create index: %w", op, err)
	}

	return &Storage{client: client, collection: coll}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "storage.mongodb.SaveUser"

	doc := UserDocument{
		Email:    email,
		Password: string(passHash),
	}

	_, err := s.collection.InsertOne(ctx, doc)
	if err != nil {
		var mongoWriteException mongo.WriteException
		if errors.As(err, &mongoWriteException) {
			for _, writeError := range mongoWriteException.WriteErrors {
				if writeError.Code == 11000 {
					return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
				}
			}
		}
		return 0, fmt.Errorf("%s: insert document: %w", op, err)
	}

	return 1, nil // Assuming you return 1 for success (MongoDB's LastInsertId equivalent)
}

func (s *Storage) GetUser(ctx context.Context, email string) (models.User, error) {
	const op = "storage.mongodb.GetUser"

	var doc UserDocument
	filter := bson.D{{Key: "email", Value: email}}

	err := s.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.User{}, storage.ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("%s: find document: %w", op, err)
	}

	return models.User{
		Email:    doc.Email,
		PassHash: []byte(doc.Password),
	}, nil
}
