package domain

import (
	"context"
	"github.com/spring-financial-group/peacock/pkg/models"
	"time"
)

type ReleaseUseCase interface {
	SaveRelease(ctx context.Context, environment string, releaseNotes []models.ReleaseNote, prSummary models.PullRequestSummary) error
	GetReleases(ctx context.Context, environment string, startTime time.Time, teams []string) ([]models.Release, error)
}

type UserReleaseTrackingUseCase interface {
	GetUnviewedReleases(ctx context.Context, userID, environment string) (*models.GetUnviewedReleasesResponse, error)
	MarkReleasesViewed(ctx context.Context, userID string, request models.MarkViewedRequest) error
	GetUserStatus(ctx context.Context, userID, environment string) (*models.GetUserStatusResponse, error)
}

type ReleaseRepository interface {
	Insert(ctx context.Context, release models.Release) error
	GetReleases(ctx context.Context, environment string, startTime time.Time, teams []string) ([]models.Release, error)
}

type UserReleaseTrackingRepository interface {
	GetUserTracking(ctx context.Context, userID, environment string) (*models.UserReleaseTracking, error)
	UpsertUserTracking(ctx context.Context, tracking *models.UserReleaseTracking) error
	MarkReleasesViewed(ctx context.Context, userID, environment string, releaseIDs []string) error
}
