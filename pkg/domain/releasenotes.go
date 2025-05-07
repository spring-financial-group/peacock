package domain

import (
	"github.com/spring-financial-group/peacock/pkg/models"
)

type ReleaseNotesUseCase interface {
	// GetReleaseNotesFromMarkdownAndTeamsInFeathers parses release notes from a markdown string attaching the corresponding teams from feathers
	GetReleaseNotesFromMarkdownAndTeamsInFeathers(markdown string, teamsInFeathers models.Teams) ([]models.ReleaseNote, error)
	// PopulateTeamsInReleaseNotes populates the teams in the release notes with the corresponding teams in feathers
	PopulateTeamsInReleaseNotes(releaseNotes []models.ReleaseNote, teamsInFeathers models.Teams) error
	// ParseReleaseNoteFromMarkdown parses release notes from a markdown string
	ParseReleaseNoteFromMarkdown(markdown string, sanitise bool) (preamble string, notes []models.ReleaseNote, err error)
	// GetMarkdownFromReleaseNotes generates a markdown string from a slice of release notes
	GetMarkdownFromReleaseNotes(notes []models.ReleaseNote) string
	// GenerateHash generates a SHA256 hash of the json of a slice of release notes
	GenerateHash(messages []models.ReleaseNote) (string, error)
	// GenerateBreakdown generates a markdown string breaking down the release notes
	GenerateBreakdown(notes []models.ReleaseNote, hash string, totalTeams int) (string, error)
	// SendReleaseNotes sends release notes to their respective teams
	SendReleaseNotes(subject string, notes []models.ReleaseNote) error
	// AppendReleaseNotesToExistingMarkdown appends release notes to an existing markdown string merging notes by team if possible.
	// If a note is not mergable, it will be appended as a new note. Order of the existing notes is preserved.
	AppendReleaseNotesToExistingMarkdown(existingMarkdown string, releaseNotesToAppend []models.ReleaseNote) (string, error)
}
