package domain

import (
	"github.com/spring-financial-group/peacock/pkg/feathers"
	"github.com/spring-financial-group/peacock/pkg/models"
)

type ReleaseNotesUseCase interface {
	// ParseNotesFromMarkdown parses release notes from a markdown string
	ParseNotesFromMarkdown(markdown string) ([]models.ReleaseNote, error)
	// GenerateHash generates a SHA256 hash of the json of a slice of release notes
	GenerateHash(messages []models.ReleaseNote) (string, error)
	// GenerateBreakdown generates a markdown string breaking down the release notes
	GenerateBreakdown(notes []models.ReleaseNote, hash string, totalTeams int) (string, error)
	// SendReleaseNotes sends release notes to their respective teams
	SendReleaseNotes(feathers *feathers.Feathers, messages []models.ReleaseNote) error
	// ValidateReleaseNotesWithFeathers checks that the relevant handlers have been registered and that the teams exist
	ValidateReleaseNotesWithFeathers(feathers *feathers.Feathers, notes []models.ReleaseNote) error
}
