package mongodb

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"ocontest/internal/db"
	"ocontest/pkg"
	"ocontest/pkg/configs"
	"ocontest/pkg/structs"
	"time"
)

// Replace the placeholder with your Atlas connection string
const timeout = time.Second
const collectionName = "problem_description"

type ProblemDescriptionRepoImp struct {
	collection *mongo.Collection
}

func NewProblemDescriptionRepo(config configs.SectionMongo) (db.ProblemDescriptionsRepo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(config.Address).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &ProblemDescriptionRepoImp{
		collection: client.Database(config.Database).Collection(collectionName),
	}, client.Ping(ctx, nil)
}

func (p ProblemDescriptionRepoImp) Save(description string) (string, error) {
	document := bson.D{
		{"description", description},
	}
	// insert into collection testc

	res, err := p.collection.InsertOne(context.Background(), document)
	if err != nil {
		return "", err
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil

}

func (p ProblemDescriptionRepoImp) Get(id string) (string, error) {
	fid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return "", err
	}

	var result structs.ProblemDescription
	err = p.collection.FindOne(context.Background(), bson.D{{"_id", fid}}, nil).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", pkg.ErrNotFound
		}
		return "", err
	}
	return result.Description, nil
}

func (p ProblemDescriptionRepoImp) Update(id, description string) error {
	fid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.D{{"_id", fid}}
	update := bson.D{{"$set", bson.D{{"description", description}}}}

	result, err := p.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return pkg.ErrNotFound
	}

	return nil
}

func (p ProblemDescriptionRepoImp) Delete(id string) error {
	fid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.D{{"_id", fid}}

	res, err := p.collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return pkg.ErrNotFound
	}

	return nil
}
