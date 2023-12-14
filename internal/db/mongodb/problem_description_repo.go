package mongodb

import (
	"context"
	"github.com/ocontest/backend/internal/db"
	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/configs"
	"github.com/ocontest/backend/pkg/structs"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// Replace the placeholder with your Atlas connection string
const timeout = time.Second

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
		collection: client.Database(config.Database).Collection("problem_description"),
	}, client.Ping(ctx, nil)
}

func (p ProblemDescriptionRepoImp) Insert(description string, testCases []string) (string, error) {
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
