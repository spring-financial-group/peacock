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

type ReleaseRepository interface {
	Insert(ctx context.Context, release models.Release) error
	GetReleases(ctx context.Context, environment string, startTime time.Time, teams []string) ([]models.Release, error)
}
