package domain

import (
	"context"
	"github.com/spring-financial-group/peacock/pkg/models"
)

type DiffsUseCase interface {
	// GenerateDiffSection renders the markdown file changes section for the changed files in a PR
	// matching the given diff configs, or an empty string when nothing matched or nothing changed
	GenerateDiffSection(ctx context.Context, owner, repoName string, prNumber int, configs models.DiffConfigs) (string, error)
}
