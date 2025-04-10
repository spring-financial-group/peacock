package releasenotesuc

import (
	"fmt"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/models"
	"github.com/spring-financial-group/peacock/pkg/msgclients/slack"
	"github.com/spring-financial-group/peacock/pkg/msgclients/webhook"
	"github.com/spring-financial-group/peacock/pkg/releasenotes/delivery/msgclients"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	devsTeam = models.Team{
		Name:        "devs",
		ContactType: models.Slack,
	}
	infraTeam = models.Team{
		Name:        "infrastructure",
		ContactType: models.Slack,
	}
	mlTeam = models.Team{
		Name:        "ml",
		ContactType: models.Slack,
	}
	productTeam = models.Team{
		Name:        "product",
		ContactType: models.Webhook,
	}
	teamWithBadContactType = models.Team{
		Name:        "teamWithBadContactType",
		ContactType: "bad",
	}
	allTeams = models.Teams{
		devsTeam,
		infraTeam,
		mlTeam,
		productTeam,
		teamWithBadContactType,
	}
)

func TestParse(t *testing.T) {
	uc := NewUseCase(&msgclients.Handler{
		Clients: map[string]domain.MessageClient{
			models.Slack:   &slack.Client{},
			models.Webhook: &webhook.Client{},
		},
	})

	testCases := []struct {
		name          string
		inputMarkdown string
		expectedNotes []models.ReleaseNote
		shouldError   bool
	}{
		{
			name:          "Passing",
			inputMarkdown: "### Notify infrastructure, devs\nTest Content\n### Notify ml\nMore Test Content",
			expectedNotes: []models.ReleaseNote{
				{
					Teams:   models.Teams{infraTeam, devsTeam},
					Content: "Test Content",
				},
				{
					Teams:   models.Teams{mlTeam},
					Content: "More Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:          "CommaSeparatedVaryingWhiteSpace",
			inputMarkdown: "### Notify infrastructure,devs, ml , product\nTest Content\n",
			expectedNotes: []models.ReleaseNote{
				{
					Teams:   models.Teams{infraTeam, devsTeam, mlTeam, productTeam},
					Content: "Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:          "HeadingsInContent",
			inputMarkdown: "### Notify infrastructure\n### Test Content\nThis is some content with headers\n#### Another different header",
			expectedNotes: []models.ReleaseNote{
				{
					Teams:   models.Teams{infraTeam},
					Content: "### Test Content\nThis is some content with headers\n#### Another different header",
				},
			},
			shouldError: false,
		},
		{
			name:          "PrefaceToMessages",
			inputMarkdown: "# Title to the PR\nSome information about the pr\n### Notify infrastructure\nTest Content",
			expectedNotes: []models.ReleaseNote{
				{
					Teams:   models.Teams{infraTeam},
					Content: "Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:          "NoInputMarkdown",
			inputMarkdown: "",
			expectedNotes: nil,
			shouldError:   false,
		},
		{
			name:          "NoMessages",
			inputMarkdown: "# Title to the PR\nSome information about the pr\n",
			expectedNotes: nil,
			shouldError:   false,
		},
		{
			name:          "NoTeams",
			inputMarkdown: "### Notify ",
			expectedNotes: nil,
			shouldError:   false,
		},
		{
			name:          "MultipleMessages",
			inputMarkdown: "### Notify infrastructure\nTest Content\n### Notify ml\nMore test content\n",
			expectedNotes: []models.ReleaseNote{
				{
					Teams:   models.Teams{infraTeam},
					Content: "Test Content",
				},
				{
					Teams:   models.Teams{mlTeam},
					Content: "More test content",
				},
			},
			shouldError: false,
		},
		{
			name:          "MultipleTeamsInOneMessage",
			inputMarkdown: "### Notify infrastructure, ml, devs\nTest Content\n",
			expectedNotes: []models.ReleaseNote{
				{
					Teams:   models.Teams{infraTeam, mlTeam, devsTeam},
					Content: "Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:          "AdditionalNewLines",
			inputMarkdown: "\n\n### Notify infrastructure\nTest Content\n\n\n",
			expectedNotes: []models.ReleaseNote{
				{
					Teams:   models.Teams{infraTeam},
					Content: "Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:          "MultiLineContent",
			inputMarkdown: "### Notify infrastructure\nThis is an example\nThat runs\nAcross multiple\nlines",
			expectedNotes: []models.ReleaseNote{
				{
					Teams:   models.Teams{infraTeam},
					Content: "This is an example\nThat runs\nAcross multiple\nlines",
				},
			},
			shouldError: false,
		},
		{
			name:          "Lists",
			inputMarkdown: "### Notify infrastructure\nHere's a list of what we've done\n\t- Fixes\n\t- Features\n\t- bugs",
			expectedNotes: []models.ReleaseNote{
				{
					Teams:   models.Teams{infraTeam},
					Content: "Here's a list of what we've done\n\t- Fixes\n\t- Features\n\t- bugs",
				},
			},
			shouldError: false,
		},
		{
			name:          "WhitespaceAfterTeamName",
			inputMarkdown: "\n### Notify infrastructure   \nTest Content",
			expectedNotes: []models.ReleaseNote{
				{
					Teams:   models.Teams{infraTeam},
					Content: "Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:          "ExtraWhitespaceBetweenTeamNames",
			inputMarkdown: "\n### Notify infrastructure   ,    ml ,   product\nTest Content",
			expectedNotes: []models.ReleaseNote{
				{
					Teams:   models.Teams{infraTeam, mlTeam, productTeam},
					Content: "Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:          "NoWhitespaceBeforeTeamName",
			inputMarkdown: "# Peacock\r\n## ReleaseNote\n### Notifyinfrastructure\nTest Content",
			expectedNotes: []models.ReleaseNote{
				{
					Teams:   models.Teams{infraTeam},
					Content: "Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:          "TeamDoesNotExist",
			inputMarkdown: "### Notify NoneExistentPeople\nTest Content",
			expectedNotes: nil,
			shouldError:   true,
		},
		{
			name:          "InvalidContactType",
			inputMarkdown: "### Notify teamWithBadContactType\nTest Content",
			expectedNotes: nil,
			shouldError:   true,
		},
		{
			name:          "RemoveBotGeneratedText",
			inputMarkdown: "### Notify infrastructure\n[//]: # (beaver-start)\nTest Content\n\n[//]: # (some-bot-content)\nAnother bit of content\n### Notify devs\nMore Test Content",
			expectedNotes: []models.ReleaseNote{
				{
					Teams:   models.Teams{infraTeam},
					Content: "Test Content\n\nAnother bit of content",
				},
				{
					Teams:   models.Teams{devsTeam},
					Content: "More Test Content",
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			actualMessages, err := uc.GetReleaseNotesFromMDAndTeams(tt.inputMarkdown, allTeams)
			if tt.shouldError {
				fmt.Println("expected error: " + err.Error())
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedNotes, actualMessages)
		})
	}
}

func TestOptions_GenerateMessageBreakdown(t *testing.T) {
	uc := NewUseCase(nil)

	testCases := []struct {
		name              string
		inputNotes        []models.ReleaseNote
		numberOfTeams     int
		expectedBreakdown string
	}{
		{
			name: "SingleMessage",
			inputNotes: []models.ReleaseNote{
				{
					Teams:   models.Teams{infraTeam},
					Content: "New release of some infrastructure\nrelated things",
				},
			},
			numberOfTeams:     1,
			expectedBreakdown: "Successfully validated 1 release note.\n\n***\nRelease Note 1 will be sent to: infrastructure\n<details>\n<summary>Release Note Breakdown</summary>\n\nNew release of some infrastructure\nrelated things\n\n</details>\n<!-- hash: ReallyGoodHash type: breakdown -->\n",
		},
		{
			name: "MultipleMessages",
			inputNotes: []models.ReleaseNote{
				{
					Teams:   models.Teams{infraTeam},
					Content: "New release of some infrastructure\nrelated things",
				},
				{
					Teams:   models.Teams{mlTeam},
					Content: "New release of some ml\nrelated things",
				},
			},
			numberOfTeams:     2,
			expectedBreakdown: "Successfully validated 2 release notes.\n\n***\nRelease Note 1 will be sent to: infrastructure\n<details>\n<summary>Release Note Breakdown</summary>\n\nNew release of some infrastructure\nrelated things\n\n</details>\n\n\n***\nRelease Note 2 will be sent to: ml\n<details>\n<summary>Release Note Breakdown</summary>\n\nNew release of some ml\nrelated things\n\n</details>\n<!-- hash: ReallyGoodHash type: breakdown -->\n",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockHash := "ReallyGoodHash"

			actualBreakdown, err := uc.GenerateBreakdown(tt.inputNotes, mockHash, tt.numberOfTeams)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBreakdown, actualBreakdown)
		})
	}
}

func TestUseCase_GetMarkdownFromReleaseNotes(t *testing.T) {
	testCases := []struct {
		name     string
		notes    []models.ReleaseNote
		expected string
	}{
		{
			name: "SingleNote",
			notes: []models.ReleaseNote{
				{
					Teams:   models.Teams{infraTeam},
					Content: "New release of some infrastructure\nrelated things",
				},
			},
			expected: "### Notify infrastructure\nNew release of some infrastructure\nrelated things",
		},
		{
			name: "MultipleNotes",
			notes: []models.ReleaseNote{
				{
					Teams:   models.Teams{infraTeam},
					Content: "New release of some infrastructure\nrelated things",
				},
				{
					Teams:   models.Teams{mlTeam},
					Content: "New release of some ml\nrelated things",
				},
			},
			expected: "### Notify infrastructure\nNew release of some infrastructure\nrelated things\n\n### Notify ml\nNew release of some ml\nrelated things",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uc := NewUseCase(nil)
			actual := uc.GetMarkdownFromReleaseNotes(tc.notes)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
