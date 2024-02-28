package mongodb

import (
	"context"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type repository struct {
	collection *mongo.Collection
}

func NewRepository(client mongo.Client) domain.ReleaseRepository {
	db := *client.Database("Peacock")
	collection := db.Collection("Release")

	return &repository{
		collection: collection,
	}
}

func (r *repository) Insert(ctx context.Context, release models.Release) error {
	_, err := r.collection.InsertOne(ctx, release)
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) GetReleases(ctx context.Context, environment string, startTime time.Time, teams []string) ([]models.Release, error) {
	filter := bson.M{"environment": environment, "createdAt": bson.M{"$gt": startTime}}
	if len(teams) > 0 {
		filter["releaseNotes.teams.name"] = bson.M{"$in": teams}
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var releases []models.Release
	if err = cursor.All(ctx, &releases); err != nil {
		return nil, err
	}

	return releases, nil
}
