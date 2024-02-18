package mongodb

import (
	"context"

	"github.com/ocontest/backend/pkg/configs"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewConn(ctx context.Context, c configs.SectionMongo) (*mongo.Client, error) {

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(c.Address).SetServerAPIOptions(serverAPI)

	return mongo.Connect(ctx, opts)
}
