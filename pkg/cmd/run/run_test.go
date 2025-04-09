package run_test

import (
	"context"
	"fmt"
	"github.com/google/go-github/v48/github"
	"github.com/spring-financial-group/peacock/pkg/cmd/run"
	"github.com/spring-financial-group/peacock/pkg/domain/mocks"
	"github.com/spring-financial-group/peacock/pkg/models"
	"github.com/spring-financial-group/peacock/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

var (
	infraTeam = models.Team{
		Name:        "infrastructure",
		APIKey:      "some-api-key",
		Addresses:   []string{},
		ContactType: models.Slack,
	}
	productTeam = models.Team{
		Name:        "product",
		APIKey:      "another-api-key",
		Addresses:   []string{},
		ContactType: models.Webhook,
	}
	allTeams = models.Teams{
		infraTeam,
		productTeam,
	}
)

func TestOptions_HaveMessagesChanged(t *testing.T) {
	// Comments returned from the GitHub API are sorted by most recent first
	testCases := []struct {
		name             string
		inputMessages    []models.ReleaseNote
		returnedComments []*github.IssueComment
		expectedHash     string
		expectedChanged  bool
	}{
		{
			name: "OneCommentSameHash",
			inputMessages: []models.ReleaseNote{
				{
					Content: "New release of some infrastructure\nrelated things",
				},
			},
			returnedComments: []*github.IssueComment{
				{
					Body: utils.NewPtr("<!-- hash: SomeReallyGoodHash type: breakdown -->"),
				},
			},
			expectedHash:    "",
			expectedChanged: false,
		},
		{
			name: "OneCommentDifferentHash",
			inputMessages: []models.ReleaseNote{
				{
					Content: "New release of some infrastructure\nrelated things",
				},
			},
			returnedComments: []*github.IssueComment{
				{
					Body: utils.NewPtr("<!-- hash: SomeOtherHash type: breakdown -->"),
				},
			},
			expectedHash:    "SomeReallyGoodHash",
			expectedChanged: true,
		},
		{
			name: "MultipleCommentsDifferentHashes",
			inputMessages: []models.ReleaseNote{
				{
					Content: "New release of some infrastructure\nrelated things",
				},
			},
			returnedComments: []*github.IssueComment{
				{
					Body: utils.NewPtr("<!-- hash: SomeOtherHash type: breakdown -->"),
				},
				{
					Body: utils.NewPtr("<!-- hash: AnotherHash type: breakdown -->"),
				},
				{
					Body: utils.NewPtr("<!-- hash: HashingHel type: breakdown -->"),
				},
				{
					Body: utils.NewPtr("<!-- hash: AllTheHashes type: breakdown -->"),
				},
			},
			expectedHash:    "SomeReallyGoodHash",
			expectedChanged: true,
		},
		{
			name: "MostRecentCommentSameHash",
			inputMessages: []models.ReleaseNote{
				{
					Content: "New release of some infrastructure\nrelated things",
				},
			},
			returnedComments: []*github.IssueComment{
				{
					Body: utils.NewPtr("<!-- hash: SomeReallyGoodHash type: breakdown -->"),
				},
				{
					Body: utils.NewPtr("<!-- hash: AnotherHash type: breakdown -->"),
				},
				{
					Body: utils.NewPtr("<!-- hash: HashingHel type: breakdown -->"),
				},
				{
					Body: utils.NewPtr("<!-- hash: AllTheHashes type: breakdown -->"),
				},
			},
			expectedHash:    "",
			expectedChanged: false,
		},
		{
			name: "MostRecentCommentDifferentHash",
			inputMessages: []models.ReleaseNote{
				{
					Content: "New release of some infrastructure\nrelated things",
				},
			},
			returnedComments: []*github.IssueComment{
				{
					Body: utils.NewPtr("<!-- hash: AnotherHash type: breakdown -->"),
				},
				{
					Body: utils.NewPtr("<!-- hash: SomeReallyGoodHash type: breakdown -->"),
				},
				{
					Body: utils.NewPtr("<!-- hash: HashingHel type: breakdown -->"),
				},
				{
					Body: utils.NewPtr("<!-- hash: AllTheHashes type: breakdown -->"),
				},
			},
			expectedHash:    "SomeReallyGoodHash",
			expectedChanged: true,
		},
		{
			name: "NoCommentsContainingMetadata",
			inputMessages: []models.ReleaseNote{
				{
					Content: "New release of some infrastructure\nrelated things",
				},
			},
			returnedComments: []*github.IssueComment{
				{
					Body: utils.NewPtr("Comment from someone else"),
				},
				{
					Body: utils.NewPtr("Comment from someone else"),
				},
				{
					Body: utils.NewPtr("Really different comment"),
				},
			},
			expectedHash:    "SomeReallyGoodHash",
			expectedChanged: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockSCM := mocks.NewSCM(t)
			mockNotesUC := mocks.NewReleaseNotesUseCase(t)

			opts := &run.Options{
				GitServerClient: mockSCM,
				NotesUC:         mockNotesUC,
			}

			mockHash := "SomeReallyGoodHash"

			mockNotesUC.On("GenerateHash", tt.inputMessages).Return(mockHash, nil)
			mockSCM.On("GetPRComments", mock.Anything, "", "", 0).Return(tt.returnedComments, nil).Once()

			actualChanged, actualHash, err := opts.HaveMessagesChanged(context.Background(), tt.inputMessages)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedChanged, actualChanged)
			assert.Equal(t, tt.expectedHash, actualHash)
		})
	}
}

func TestOptions_Run(t *testing.T) {
	mockSCM := mocks.NewSCM(t)
	mockGitClient := mocks.NewGit(t)
	mockNotesUC := mocks.NewReleaseNotesUseCase(t)

	testCases := []struct {
		name        string
		opts        *run.Options
		prBody      *string
		shouldError bool
	}{
		{
			name: "NonDryRun",
			opts: &run.Options{
				PRNumber:          1,
				GitServerURL:      "https://github.com",
				GitHubToken:       "testGitToken",
				RepoOwner:         "spring-financial-group",
				RepoName:          "peacock",
				DryRun:            false,
				CommentValidation: true,
				SlackToken:        "testSlackToken",
				Git:               mockGitClient,
				GitServerClient:   mockSCM,
				NotesUC:           mockNotesUC,
				Feathers: &models.Feathers{
					Teams: allTeams,
				},
			},
			prBody: utils.NewPtr("# Peacock\r\n## ReleaseNote\n### Notify infrastructure\nTest Content"),
		},
		{
			name: "DryRun",
			opts: &run.Options{
				PRNumber:          1,
				GitServerURL:      "https://github.com",
				GitHubToken:       "testGitToken",
				RepoOwner:         "spring-financial-group",
				RepoName:          "peacock",
				DryRun:            true,
				CommentValidation: true,
				SlackToken:        "testSlackToken",
				Git:               mockGitClient,
				GitServerClient:   mockSCM,
				NotesUC:           mockNotesUC,
				Feathers: &models.Feathers{
					Teams: allTeams,
				},
			},
			prBody: utils.NewPtr("# Peacock\r\n## ReleaseNote\n### Notify infrastructure\nTest Content"),
		},
	}

	for _, tt := range testCases {
		mockNotes := []models.ReleaseNote{
			{
				Teams: models.Teams{
					infraTeam,
				},
				Content: "Test Content",
			},
		}

		mockHash := "SomeReallyGoodHash"
		mockBreakdown := "This is the message breakdown"

		if tt.opts.DryRun {
			mockNotesUC.On("GenerateHash", mockNotes).Return(mockHash, nil)
			mockNotesUC.On("GenerateBreakdown", mockNotes, mockHash, len(allTeams)).Return(mockBreakdown, nil)

			mockSCM.On("GetPullRequestBodyFromPRNumber", mock.Anything, "spring-financial-group", "peacock", 1).Return(tt.prBody, nil).Once()
			mockSCM.On("CommentOnPR", mock.Anything, "spring-financial-group", "peacock", 1, mock.AnythingOfType("string")).Return(nil).Once()
			mockSCM.On("GetPRComments", mock.Anything, "spring-financial-group", "peacock", 1).Return(nil, nil)
		} else {
			mockGitClient.On("GetLatestCommitSHA", "").Return("SHA", nil)
			mockSCM.On("GetPullRequestBodyFromCommit", mock.Anything, "spring-financial-group", "peacock", "SHA").Return(tt.prBody, nil).Once()
		}

		mockNotesUC.On("GetReleaseNotesFromMDAndTeams", *tt.prBody, allTeams, false).Return(mockNotes, nil)

		if !tt.opts.DryRun {
			mockNotesUC.On("SendReleaseNotes", "New Release Notes for peacock", mockNotes).Return(nil).Once()
		}

		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Run()
			if tt.shouldError {
				fmt.Println("expected error: " + err.Error())
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
