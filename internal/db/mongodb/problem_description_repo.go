package mongodb

import (
	"context"
	"github.com/pkg/errors"
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

func (p ProblemDescriptionRepoImp) Save(description string, testCases []string) (string, error) {
	document := bson.D{
		{"description", description},
		{"testcases", testCases},
	}

	res, err := p.collection.InsertOne(context.Background(), document)
	if err != nil {
		return "", err
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil

}

func (p ProblemDescriptionRepoImp) Get(id string) (structs.ProblemDescription, error) {
	fid, err := primitive.ObjectIDFromHex(id)
	var ans structs.ProblemDescription
	if err != nil {
		return ans, err
	}

	err = p.collection.FindOne(context.Background(), bson.D{{"_id", fid}}, nil).Decode(&ans)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ans, pkg.ErrNotFound
		}
		return ans, err
	}
	return ans, nil
}

func (p ProblemDescriptionRepoImp) AddTestcase(ctx context.Context, id string, testCase string) error {

	fid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.Wrap(err, "couldn't gen id from hex")
	}

	change := bson.M{"$push": bson.M{"testcases": testCase}}

	_, err = p.collection.UpdateOne(ctx, bson.D{{
		"_id", fid,
	}}, change, nil)
	return err
}
