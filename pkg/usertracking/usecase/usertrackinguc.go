package usecase

import (
	"context"
	"time"

	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/models"
)

type userReleaseTrackingUseCase struct {
	userTrackingRepo domain.UserReleaseTrackingRepository
	releaseRepo      domain.ReleaseRepository
	releaseNotesUC   domain.ReleaseNotesUseCase
}

func NewUserReleaseTrackingUseCase(
	userTrackingRepo domain.UserReleaseTrackingRepository,
	releaseRepo domain.ReleaseRepository,
	releaseNotesUC domain.ReleaseNotesUseCase,
) domain.UserReleaseTrackingUseCase {
	return &userReleaseTrackingUseCase{
		userTrackingRepo: userTrackingRepo,
		releaseRepo:      releaseRepo,
		releaseNotesUC:   releaseNotesUC,
	}
}

func (uc *userReleaseTrackingUseCase) GetUnviewedReleases(ctx context.Context, userID, environment string) (*models.GetUnviewedReleasesResponse, error) {
	// Get user's tracking data
	tracking, err := uc.userTrackingRepo.GetUserTracking(ctx, userID, environment)
	if err != nil {
		return nil, err
	}

	// Get all releases for the environment
	// Using a time far in the past to get all releases, could be optimized based on requirements
	startTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	allReleases, err := uc.releaseRepo.GetReleases(ctx, environment, startTime, nil)
	if err != nil {
		return nil, err
	}

	// Create a map of viewed release IDs for quick lookup
	viewedMap := make(map[string]bool)
	if tracking != nil {
		for _, viewed := range tracking.ViewedReleases {
			viewedMap[viewed.ReleaseID] = true
		}
	}

	// Filter out viewed releases
	var unviewedReleases []models.Release
	for _, release := range allReleases {
		// Generate release ID using the hash of release notes
		releaseID, err := uc.releaseNotesUC.GenerateHash(release.ReleaseNotes)
		if err != nil {
			return nil, err
		}

		if !viewedMap[releaseID] {
			unviewedReleases = append(unviewedReleases, release)
		}
	}

	return &models.GetUnviewedReleasesResponse{
		Releases:   unviewedReleases,
		TotalCount: len(unviewedReleases),
	}, nil
}

func (uc *userReleaseTrackingUseCase) MarkReleasesViewed(ctx context.Context, userID string, request models.MarkViewedRequest) error {
	return uc.userTrackingRepo.MarkReleasesViewed(ctx, userID, request.Environment, request.ReleaseIDs)
}

func (uc *userReleaseTrackingUseCase) GetUserStatus(ctx context.Context, userID, environment string) (*models.GetUserStatusResponse, error) {
	tracking, err := uc.userTrackingRepo.GetUserTracking(ctx, userID, environment)
	if err != nil {
		return nil, err
	}

	if tracking == nil {
		return &models.GetUserStatusResponse{
			ViewedReleases: []models.ViewedRelease{},
			LastChecked:    time.Time{},
		}, nil
	}

	return &models.GetUserStatusResponse{
		ViewedReleases: tracking.ViewedReleases,
		LastChecked:    tracking.LastChecked,
	}, nil
}
