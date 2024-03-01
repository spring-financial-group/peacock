package usecase

import (
	"context"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/models"
	"time"
)

type useCase struct {
	repository domain.ReleaseRepository
}

func NewUseCase(repository domain.ReleaseRepository) domain.ReleaseUseCase {
	return &useCase{
		repository: repository,
	}
}

func (uc *useCase) SaveRelease(ctx context.Context, environment string, releaseNotes []models.ReleaseNote, pr models.PullRequestSummary) error {
	release := models.Release{
		CreatedAt:    time.Now(),
		ReleaseNotes: releaseNotes,
		Environment:  environment,
		PullRequest:  pr,
	}

	err := uc.repository.Insert(ctx, release)
	if err != nil {
		return err
	}

	return nil
}

func (uc *useCase) GetReleases(ctx context.Context, environment string, startTime time.Time, teams []string) ([]models.Release, error) {
	releases, err := uc.repository.GetReleases(ctx, environment, startTime, teams)
	if err != nil {
		return nil, err
	}

	return releases, nil
}
