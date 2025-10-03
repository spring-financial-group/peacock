package mongodb

import (
	"context"
	"time"

	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type userReleaseTrackingRepository struct {
	collection *mongo.Collection
}

func NewUserReleaseTrackingRepository(client mongo.Client) domain.UserReleaseTrackingRepository {
	db := *client.Database("Peacock")
	collection := db.Collection("user_release_tracking")

	return &userReleaseTrackingRepository{
		collection: collection,
	}
}

func (r *userReleaseTrackingRepository) GetUserTracking(ctx context.Context, userID, environment string) (*models.UserReleaseTracking, error) {
	filter := bson.M{"userId": userID, "environment": environment}

	var tracking models.UserReleaseTracking
	err := r.collection.FindOne(ctx, filter).Decode(&tracking)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &tracking, nil
}

func (r *userReleaseTrackingRepository) UpsertUserTracking(ctx context.Context, tracking *models.UserReleaseTracking) error {
	filter := bson.M{"userId": tracking.UserID, "environment": tracking.Environment}

	update := bson.M{
		"$set": bson.M{
			"userId":         tracking.UserID,
			"viewedReleases": tracking.ViewedReleases,
			"lastChecked":    tracking.LastChecked,
			"environment":    tracking.Environment,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)

	return err
}

func (r *userReleaseTrackingRepository) MarkReleasesViewed(ctx context.Context, userID, environment string, releaseIDs []string) error {
	filter := bson.M{"userId": userID, "environment": environment}

	viewedReleases := make([]models.ViewedRelease, len(releaseIDs))
	now := time.Now()
	for i, releaseID := range releaseIDs {
		viewedReleases[i] = models.ViewedRelease{
			ReleaseID: releaseID,
			ViewedAt:  now,
		}
	}

	update := bson.M{
		"$addToSet": bson.M{
			"viewedReleases": bson.M{"$each": viewedReleases},
		},
		"$set": bson.M{
			"lastChecked": now,
		},
		"$setOnInsert": bson.M{
			"userId":      userID,
			"environment": environment,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)

	return err
}
