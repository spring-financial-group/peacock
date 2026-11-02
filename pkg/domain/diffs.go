package domain

import (
	"context"
	"github.com/spring-financial-group/peacock/pkg/models"
)

type DiffsUseCase interface {
	// GenerateDiffSection renders the markdown file changes for the changed files in a PR matching
	// the given diff configs, keyed by the feather team each diff attaches to (empty key when none).
	// Returns an empty map when nothing matched or nothing changed
	GenerateDiffSection(ctx context.Context, owner, repoName string, prNumber int, configs models.DiffConfigs) (models.DiffSections, error)
}
