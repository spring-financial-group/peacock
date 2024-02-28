package usecase

import (
	"context"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/models"
	"time"
)

type useCase struct {
	repository domain.ReleaseRepository
	scm        domain.SCM
}

func NewUseCase(repository domain.ReleaseRepository) domain.ReleaseUseCase {
	return &useCase{
		repository: repository,
	}
}

func (uc *useCase) SaveRelease(ctx context.Context, environment string, releaseNotes []models.ReleaseNote) error {
	release := models.Release{
		CreatedAt:    time.Now(),
		ReleaseNotes: releaseNotes,
		Environment:  environment,
	}

	err := uc.repository.Insert(ctx, release)
	if err != nil {
		return err
	}

	return nil
}

func (uc *useCase) GetReleasesAfterDate(ctx context.Context, environment string, startTime time.Time) ([]models.Release, error) {
	releases, err := uc.repository.GetReleasesAfterDate(ctx, environment, startTime)
	if err != nil {
		return nil, err
	}

	return releases, nil
}
