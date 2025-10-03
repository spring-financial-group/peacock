package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/spring-financial-group/peacock/pkg/domain/mocks"
	"github.com/spring-financial-group/peacock/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserReleaseTrackingUseCase_GetUnviewedReleases(t *testing.T) {
	// Setup mocks
	userTrackingRepo := &mocks.UserReleaseTrackingRepository{}
	releaseRepo := &mocks.ReleaseRepository{}
	releaseNotesUC := &mocks.ReleaseNotesUseCase{}

	// Setup test data
	userID := "testuser"
	environment := "production"
	releaseNotes := []models.ReleaseNote{
		{
			Teams:   []models.Team{{Name: "team1"}},
			Content: "Test release note",
		},
	}

	releases := []models.Release{
		{
			CreatedAt:    time.Now(),
			ReleaseNotes: releaseNotes,
			Environment:  environment,
		},
	}

	expectedHash := "testhash123"

	// Setup expectations
	userTrackingRepo.On("GetUserTracking", mock.Anything, userID, environment).Return(nil, nil)
	releaseRepo.On("GetReleases", mock.Anything, environment, mock.AnythingOfType("time.Time"), mock.Anything).Return(releases, nil)
	releaseNotesUC.On("GenerateHash", releaseNotes).Return(expectedHash, nil)

	// Create use case
	uc := NewUserReleaseTrackingUseCase(userTrackingRepo, releaseRepo, releaseNotesUC)

	// Execute test
	result, err := uc.GetUnviewedReleases(context.Background(), userID, environment)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.TotalCount)
	assert.Len(t, result.Releases, 1)

	// Verify all mocks were called
	userTrackingRepo.AssertExpectations(t)
	releaseRepo.AssertExpectations(t)
	releaseNotesUC.AssertExpectations(t)
}

func TestUserReleaseTrackingUseCase_MarkReleasesViewed(t *testing.T) {
	// Setup mocks
	userTrackingRepo := &mocks.UserReleaseTrackingRepository{}
	releaseRepo := &mocks.ReleaseRepository{}
	releaseNotesUC := &mocks.ReleaseNotesUseCase{}

	// Setup test data
	userID := "testuser"
	request := models.MarkViewedRequest{
		ReleaseIDs:  []string{"release1", "release2"},
		Environment: "production",
	}

	// Setup expectations
	userTrackingRepo.On("MarkReleasesViewed", mock.Anything, userID, request.Environment, request.ReleaseIDs).Return(nil)

	// Create use case
	uc := NewUserReleaseTrackingUseCase(userTrackingRepo, releaseRepo, releaseNotesUC)

	// Execute test
	err := uc.MarkReleasesViewed(context.Background(), userID, request)

	// Assertions
	assert.NoError(t, err)

	// Verify all mocks were called
	userTrackingRepo.AssertExpectations(t)
}
