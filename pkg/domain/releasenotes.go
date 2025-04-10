package domain

import (
	"github.com/spring-financial-group/peacock/pkg/models"
)

type ReleaseNotesUseCase interface {
	// GetReleaseNotesFromMDAndTeams parses release notes from a markdown string attaching the corresponding teams
	GetReleaseNotesFromMDAndTeams(markdown string, teamsInFeathers models.Teams) ([]models.ReleaseNote, error)
	// GetMarkdownFromReleaseNotes generates a markdown string from a slice of release notes
	GetMarkdownFromReleaseNotes(notes []models.ReleaseNote) string
	// GenerateHash generates a SHA256 hash of the json of a slice of release notes
	GenerateHash(messages []models.ReleaseNote) (string, error)
	// GenerateBreakdown generates a markdown string breaking down the release notes
	GenerateBreakdown(notes []models.ReleaseNote, hash string, totalTeams int) (string, error)
	// SendReleaseNotes sends release notes to their respective teams
	SendReleaseNotes(subject string, notes []models.ReleaseNote) error
}
