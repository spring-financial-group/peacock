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
			mockSCM.On("GetPRComments", mock.AnythingOfType("*context.emptyCtx")).Return(tt.returnedComments, nil).Once()

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
					Teams: []models.Team{
						{
							Name:        "infrastructure",
							ContactType: models.Slack,
							Addresses:   []string{"TestAdd", "TestAdd2"},
						},
					},
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
					Teams: []models.Team{
						{
							Name:        "infrastructure",
							ContactType: models.Slack,
							Addresses:   []string{"TestAdd", "TestAdd2"},
						},
					},
				},
			},
			prBody: utils.NewPtr("# Peacock\r\n## ReleaseNote\n### Notify infrastructure\nTest Content"),
		},
	}

	for _, tt := range testCases {
		mockNotes := []models.ReleaseNote{
			{
				TeamNames: []string{"infrastructure"},
				Content:   "Test Content",
			},
		}

		mockHash := "SomeReallyGoodHash"
		mockBreakdown := "This is the message breakdown"

		if tt.opts.DryRun {
			mockNotesUC.On("GenerateHash", mockNotes).Return(mockHash, nil)
			mockNotesUC.On("GenerateBreakdown", mockNotes, mockHash, 1).Return(mockBreakdown, nil)

			mockSCM.On("GetPullRequestBodyFromPRNumber", mock.AnythingOfType("*context.emptyCtx")).Return(tt.prBody, nil).Once()
			mockSCM.On("CommentOnPR", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("string")).Return(nil).Once()
			mockSCM.On("GetPRComments", mock.AnythingOfType("*context.emptyCtx")).Return(nil, nil)
		} else {
			mockGitClient.On("GetLatestCommitSHA", "").Return("SHA", nil)
			mockSCM.On("GetPullRequestBodyFromCommit", mock.AnythingOfType("*context.emptyCtx"), "SHA").Return(tt.prBody, nil).Once()
		}

		mockNotesUC.On("ParseNotesFromMarkdown", *tt.prBody).Return(mockNotes, nil)

		mockNotesUC.On("ValidateReleaseNotesWithFeathers", tt.opts.Feathers, mockNotes).Return(nil)

		if !tt.opts.DryRun {
			mockNotesUC.On("SendReleaseNotes", tt.opts.Feathers, mockNotes).Return(nil).Once()
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
