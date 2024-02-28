package domain

import (
	"context"
	"github.com/spring-financial-group/peacock/pkg/models"
	"time"
)

type ReleaseUseCase interface {
	SaveRelease(ctx context.Context, environment string, releaseNotes []models.ReleaseNote) error
	GetReleasesAfterDate(ctx context.Context, environment string, startTime time.Time) ([]models.Release, error)
}

type ReleaseRepository interface {
	Insert(ctx context.Context, release models.Release) error
	GetReleasesAfterDate(ctx context.Context, environment string, startTime time.Time) ([]models.Release, error)
}
