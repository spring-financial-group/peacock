package run_test

import (
	"fmt"
	"github.com/spring-financial-group/peacock/pkg/cmd/run"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/domain/mocks"
	"github.com/spring-financial-group/peacock/pkg/feathers"
	"github.com/spring-financial-group/peacock/pkg/handlers"
	"github.com/spring-financial-group/peacock/pkg/message"
	"github.com/spring-financial-group/peacock/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestOptions_Run(t *testing.T) {
	mockGitServer := mocks.NewGitServer(t)
	mockGitClient := mocks.NewGit(t)
	mockSlackHander := mocks.NewMessageHandler(t)

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
				GitServerClient:   mockGitServer,
				Handlers:          map[string]domain.MessageHandler{handlers.Slack: mockSlackHander},
				Config: &feathers.Feathers{
					Teams: []feathers.Team{
						{
							Name:        "infrastructure",
							ContactType: handlers.Slack,
							Addresses:   []string{"TestAdd", "TestAdd2"},
						},
					},
				},
			},
			prBody: utils.NewPtr("# Peacock\r\n## Message\n### Notify infrastructure\nTest Content"),
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
				GitServerClient:   mockGitServer,
				Handlers:          map[string]domain.MessageHandler{handlers.Slack: mockSlackHander},
				Config: &feathers.Feathers{
					Teams: []feathers.Team{
						{
							Name:        "infrastructure",
							ContactType: handlers.Slack,
							Addresses:   []string{"TestAdd", "TestAdd2"},
						},
					},
				},
			},
			prBody: utils.NewPtr("# Peacock\r\n## Message\n### Notify infrastructure\nTest Content"),
		},
	}

	for _, tt := range testCases {
		if tt.opts.DryRun {
			mockGitServer.On("GetPullRequestBodyFromPRNumber", mock.AnythingOfType("*context.emptyCtx"), tt.opts.PRNumber).Return(tt.prBody, nil).Once()
			mockGitServer.On("CommentOnPR", mock.AnythingOfType("*context.emptyCtx"), tt.opts.PRNumber, mock.AnythingOfType("string")).Return(nil).Once()
		} else {
			mockGitClient.On("GetLatestCommitSHA").Return("SHA", nil)
			mockGitServer.On("GetPullRequestBodyFromCommit", mock.AnythingOfType("*context.emptyCtx"), "SHA").Return(tt.prBody, nil).Once()
		}

		for _, team := range tt.opts.Config.Teams {
			if !tt.opts.DryRun {
				mockSlackHander.On("Send", "Test Content", "New Release Notes for peacock", team.Addresses).Return(nil).Once()
			}
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

func TestOptions_GenerateMessageBreakdown(t *testing.T) {
	testCases := []struct {
		name              string
		opts              *run.Options
		inputMessages     []message.Message
		expectedBreakdown string
	}{
		{
			name: "OneMessage",
			opts: &run.Options{
				Config: &feathers.Feathers{
					Teams: []feathers.Team{
						{Name: "infrastructure"},
					},
				},
			},
			inputMessages: []message.Message{
				{
					TeamNames: []string{"infrastructure"},
					Content:   "New release of some infrastructure\nrelated things",
				},
			},
			expectedBreakdown: "[Peacock Validation] Successfully parsed 1 message(s).\n---\nMessage 1 will be sent to: infrastructure\n<details open>\n<summary>Message Breakdown</summary>\nNew release of some infrastructure\nrelated things\n</details>",
		},
		{
			name: "MultipleMessages&MultipleTeams",
			opts: &run.Options{
				Config: &feathers.Feathers{
					Teams: []feathers.Team{
						{Name: "infrastructure"},
						{Name: "ml"},
					},
				},
			},
			inputMessages: []message.Message{
				{
					TeamNames: []string{"infrastructure"},
					Content:   "New release of some infrastructure\nrelated things",
				},
				{
					TeamNames: []string{"ml"},
					Content:   "New release of some ml\nrelated things",
				},
			},
			expectedBreakdown: "[Peacock Validation] Successfully parsed 2 message(s).\n---\n\nMessage 1 will be sent to: infrastructure\n<details open>\n<summary>Message Breakdown</summary>\nNew release of some infrastructure\nrelated things\n</details>\nMessage 2 will be sent to: ml\n<details open>\n<summary>Message Breakdown</summary>\nNew release of some ml\nrelated things\n</details>",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			actualBreakdown, err := tt.opts.GenerateMessageBreakdown(tt.inputMessages)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBreakdown, actualBreakdown)
		})
	}
}

func TestOptions_ValidateMessagesWithConfig(t *testing.T) {
	testCases := []struct {
		name          string
		opts          *run.Options
		inputMessages []message.Message
		shouldError   bool
	}{
		{
			name: "Passing",
			opts: &run.Options{
				Handlers: map[string]domain.MessageHandler{handlers.Slack: mocks.NewMessageHandler(t)},
				Config: &feathers.Feathers{
					Teams: []feathers.Team{
						{Name: "infrastructure", ContactType: handlers.Slack},
					},
				},
			},
			inputMessages: []message.Message{
				{
					TeamNames: []string{"infrastructure"},
					Content:   "some content",
				},
			},
			shouldError: false,
		},
		{
			name: "TeamDoesNotExist",
			opts: &run.Options{
				Handlers: map[string]domain.MessageHandler{handlers.Slack: mocks.NewMessageHandler(t)},
				Config: &feathers.Feathers{
					Teams: []feathers.Team{
						{Name: "infrastructure", ContactType: handlers.Slack},
					},
				},
			},
			inputMessages: []message.Message{
				{
					TeamNames: []string{"ml"},
					Content:   "some content",
				},
			},
			shouldError: true,
		},
		{
			name: "HandlerDoesNotExist",
			opts: &run.Options{
				Handlers: map[string]domain.MessageHandler{},
				Config: &feathers.Feathers{
					Teams: []feathers.Team{
						{Name: "infrastructure", ContactType: handlers.Slack},
					},
				},
			},
			inputMessages: []message.Message{
				{
					TeamNames: []string{"infrastructure"},
					Content:   "some content",
				},
			},
			shouldError: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.ValidateMessagesWithConfig(tt.inputMessages)
			if tt.shouldError {
				fmt.Println("expected error: " + err.Error())
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
