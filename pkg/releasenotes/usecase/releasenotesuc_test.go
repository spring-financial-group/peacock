package releasenotesuc

import (
	"fmt"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/models"
	"github.com/spring-financial-group/peacock/pkg/msgclients/slack"
	"github.com/spring-financial-group/peacock/pkg/releasenotes/delivery/msgclients"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParse(t *testing.T) {
	uc := NewUseCase(nil)

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
					TeamNames: []string{"infrastructure", "devs"},
					Content:   "Test Content",
				},
				{
					TeamNames: []string{"ml"},
					Content:   "More Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:          "CommaSeperatedVaryingWhiteSpace",
			inputMarkdown: "### Notify infrastructure,devs, ml , product\nTest Content\n",
			expectedNotes: []models.ReleaseNote{
				{
					TeamNames: []string{"infrastructure", "devs", "ml", "product"},
					Content:   "Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:          "HeadingsInContent",
			inputMarkdown: "### Notify infrastructure\n### Test Content\nThis is some content with headers\n#### Another different header",
			expectedNotes: []models.ReleaseNote{
				{
					TeamNames: []string{"infrastructure"},
					Content:   "### Test Content\nThis is some content with headers\n#### Another different header",
				},
			},
			shouldError: false,
		},
		{
			name:          "PrefaceToMessages",
			inputMarkdown: "# Title to the PR\nSome information about the pr\n### Notify infrastructure\nTest Content",
			expectedNotes: []models.ReleaseNote{
				{
					TeamNames: []string{"infrastructure"},
					Content:   "Test Content",
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
			inputMarkdown: "### Notify infrastructure\nTest Content\n### Notify ML\nMore test content\n",
			expectedNotes: []models.ReleaseNote{
				{
					TeamNames: []string{"infrastructure"},
					Content:   "Test Content",
				},
				{
					TeamNames: []string{"ML"},
					Content:   "More test content",
				},
			},
			shouldError: false,
		},
		{
			name:          "MultipleTeamsInOneMessage",
			inputMarkdown: "### Notify infrastructure, ml, allDevs\nTest Content\n",
			expectedNotes: []models.ReleaseNote{
				{
					TeamNames: []string{"infrastructure", "ml", "allDevs"},
					Content:   "Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:          "AdditionalNewLines",
			inputMarkdown: "\n\n### Notify infrastructure\nTest Content\n\n\n",
			expectedNotes: []models.ReleaseNote{
				{
					TeamNames: []string{"infrastructure"},
					Content:   "Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:          "MultiLineContent",
			inputMarkdown: "### Notify infrastructure\nThis is an example\nThat runs\nAcross multiple\nlines",
			expectedNotes: []models.ReleaseNote{
				{
					TeamNames: []string{"infrastructure"},
					Content:   "This is an example\nThat runs\nAcross multiple\nlines",
				},
			},
			shouldError: false,
		},
		{
			name:          "Lists",
			inputMarkdown: "### Notify infrastructure\nHere's a list of what we've done\n\t- Fixes\n\t- Features\n\t- bugs",
			expectedNotes: []models.ReleaseNote{
				{
					TeamNames: []string{"infrastructure"},
					Content:   "Here's a list of what we've done\n\t- Fixes\n\t- Features\n\t- bugs",
				},
			},
			shouldError: false,
		},
		{
			name:          "WhitespaceAfterTeamName",
			inputMarkdown: "\n### Notify infrastructure   \nTest Content",
			expectedNotes: []models.ReleaseNote{
				{
					TeamNames: []string{"infrastructure"},
					Content:   "Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:          "ExtraWhitespaceBetweenTeamNames",
			inputMarkdown: "\n### Notify infrastructure   ,    ml ,   product\nTest Content",
			expectedNotes: []models.ReleaseNote{
				{
					TeamNames: []string{"infrastructure", "ml", "product"},
					Content:   "Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:          "NoWhitespaceBeforeTeamName",
			inputMarkdown: "# Peacock\r\n## ReleaseNote\n### Notifyinfrastructure\nTest Content",
			expectedNotes: []models.ReleaseNote{
				{
					TeamNames: []string{"infrastructure"},
					Content:   "Test Content",
				},
			},
			shouldError: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			actualMessages, err := uc.GetReleaseNotesFromMDAndTeams(tt.inputMarkdown)
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
					TeamNames: []string{"infrastructure"},
					Content:   "New release of some infrastructure\nrelated things",
				},
			},
			numberOfTeams:     1,
			expectedBreakdown: "Successfully validated 1 release note.\n\n***\nRelease Note 1 will be sent to: infrastructure\n<details>\n<summary>Release Note Breakdown</summary>\n\nNew release of some infrastructure\nrelated things\n\n</details>\n<!-- hash: ReallyGoodHash type: breakdown -->\n",
		},
		{
			name: "MultipleMessages",
			inputNotes: []models.ReleaseNote{
				{
					TeamNames: []string{"infrastructure"},
					Content:   "New release of some infrastructure\nrelated things",
				},
				{
					TeamNames: []string{"ml"},
					Content:   "New release of some ml\nrelated things",
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

func TestOptions_ValidateMessagesWithConfig(t *testing.T) {
	testCases := []struct {
		name              string
		msgClientsHandler msgclients.Handler
		inputNotes        []models.ReleaseNote
		inputFeathers     *models.Feathers
		shouldError       bool
	}{
		{
			name: "Passing",
			msgClientsHandler: msgclients.Handler{
				Clients: map[string]domain.MessageClient{
					models.Slack: slack.NewClient(""),
				},
			},
			inputFeathers: &models.Feathers{
				Teams: []models.Team{
					{Name: "infrastructure", ContactType: models.Slack},
				},
			},
			inputNotes: []models.ReleaseNote{
				{
					TeamNames: []string{"infrastructure"},
					Content:   "some content",
				},
			},
			shouldError: false,
		},
		{
			name: "TeamDoesNotExist",
			msgClientsHandler: msgclients.Handler{
				Clients: map[string]domain.MessageClient{
					models.Slack: slack.NewClient(""),
				},
			},
			inputFeathers: &models.Feathers{
				Teams: []models.Team{
					{Name: "infrastructure", ContactType: models.Slack},
				},
			},
			inputNotes: []models.ReleaseNote{
				{
					TeamNames: []string{"ml"},
					Content:   "some content",
				},
			},
			shouldError: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewUseCase(&tt.msgClientsHandler)

			err := uc.ValidateReleaseNotesWithFeathers(tt.inputFeathers, tt.inputNotes)
			if tt.shouldError {
				fmt.Println("expected error: " + err.Error())
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
