package run_test

import (
	"context"
	"fmt"
	"github.com/google/go-github/v47/github"
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

func TestOptions_HaveMessagesChanged(t *testing.T) {
	mockGitServer := mocks.NewGitServer(t)
	// Comments returned from the GitHub API are sorted by most recent first
	testCases := []struct {
		name             string
		inputMessages    []message.Message
		returnedComments []*github.IssueComment
		expectedHash     string
		expectedChanged  bool
	}{
		{
			name: "OneCommentSameHash",
			inputMessages: []message.Message{
				{
					Content: "New release of some infrastructure\nrelated things",
				},
			},
			returnedComments: []*github.IssueComment{
				{
					Body: utils.NewPtr("<!-- hash: d88cd4f055916a0a0cda7d44644750bf6008db30bbfc4ed8ee1dc8888aa817d9 -->"),
				},
			},
			expectedHash:    "",
			expectedChanged: false,
		},
		{
			name: "OneCommentDifferentHash",
			inputMessages: []message.Message{
				{
					Content: "New release of some infrastructure\nrelated things",
				},
			},
			returnedComments: []*github.IssueComment{
				{
					Body: utils.NewPtr("<!-- hash: SomeOtherHash -->"),
				},
			},
			expectedHash:    "d88cd4f055916a0a0cda7d44644750bf6008db30bbfc4ed8ee1dc8888aa817d9",
			expectedChanged: true,
		},
		{
			name: "MultipleCommentsDifferentHashes",
			inputMessages: []message.Message{
				{
					Content: "New release of some infrastructure\nrelated things",
				},
			},
			returnedComments: []*github.IssueComment{
				{
					Body: utils.NewPtr("<!-- hash: SomeOtherHash -->"),
				},
				{
					Body: utils.NewPtr("<!-- hash: AnotherHash -->"),
				},
				{
					Body: utils.NewPtr("<!-- hash: HashingHel -->"),
				},
				{
					Body: utils.NewPtr("<!-- hash: AllTheHashes -->"),
				},
			},
			expectedHash:    "d88cd4f055916a0a0cda7d44644750bf6008db30bbfc4ed8ee1dc8888aa817d9",
			expectedChanged: true,
		},
		{
			name: "MostRecentCommentSameHash",
			inputMessages: []message.Message{
				{
					Content: "New release of some infrastructure\nrelated things",
				},
			},
			returnedComments: []*github.IssueComment{
				{
					Body: utils.NewPtr("<!-- hash: d88cd4f055916a0a0cda7d44644750bf6008db30bbfc4ed8ee1dc8888aa817d9 -->"),
				},
				{
					Body: utils.NewPtr("<!-- hash: AnotherHash -->"),
				},
				{
					Body: utils.NewPtr("<!-- hash: HashingHel -->"),
				},
				{
					Body: utils.NewPtr("<!-- hash: AllTheHashes -->"),
				},
			},
			expectedHash:    "",
			expectedChanged: false,
		},
		{
			name: "MostRecentCommentDifferentHash",
			inputMessages: []message.Message{
				{
					Content: "New release of some infrastructure\nrelated things",
				},
			},
			returnedComments: []*github.IssueComment{
				{
					Body: utils.NewPtr("<!-- hash: AnotherHash -->"),
				},
				{
					Body: utils.NewPtr("<!-- hash: d88cd4f055916a0a0cda7d44644750bf6008db30bbfc4ed8ee1dc8888aa817d9 -->"),
				},
				{
					Body: utils.NewPtr("<!-- hash: HashingHel -->"),
				},
				{
					Body: utils.NewPtr("<!-- hash: AllTheHashes -->"),
				},
			},
			expectedHash:    "d88cd4f055916a0a0cda7d44644750bf6008db30bbfc4ed8ee1dc8888aa817d9",
			expectedChanged: true,
		},
		{
			name: "NoCommentsContainingMetadata",
			inputMessages: []message.Message{
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
			expectedHash:    "d88cd4f055916a0a0cda7d44644750bf6008db30bbfc4ed8ee1dc8888aa817d9",
			expectedChanged: true,
		},
	}

	opts := &run.Options{
		GitServerClient: mockGitServer,
	}

	for _, tt := range testCases {
		mockGitServer.On("GetPRComments", mock.AnythingOfType("*context.emptyCtx"), "", "", opts.PRNumber).Return(tt.returnedComments, nil).Once()

		t.Run(tt.name, func(t *testing.T) {
			actualChanged, actualHash, err := opts.HaveMessagesChanged(context.Background(), tt.inputMessages)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedChanged, actualChanged)
			assert.Equal(t, tt.expectedHash, actualHash)
		})
	}
}

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
			mockGitServer.On("GetPullRequestBodyFromPRNumber", mock.AnythingOfType("*context.emptyCtx"), tt.opts.RepoOwner, tt.opts.RepoName, tt.opts.PRNumber).Return(tt.prBody, nil).Once()
			mockGitServer.On("CommentOnPR", mock.AnythingOfType("*context.emptyCtx"), tt.opts.RepoOwner, tt.opts.RepoName, tt.opts.PRNumber, mock.AnythingOfType("string")).Return(nil).Once()
			mockGitServer.On("GetPRComments", mock.AnythingOfType("*context.emptyCtx"), tt.opts.RepoOwner, tt.opts.RepoName, tt.opts.PRNumber).Return(nil, nil)
		} else {
			mockGitClient.On("GetLatestCommitSHA", "").Return("SHA", nil)
			mockGitServer.On("GetPullRequestBodyFromCommit", mock.AnythingOfType("*context.emptyCtx"), tt.opts.RepoOwner, tt.opts.RepoName, "SHA").Return(tt.prBody, nil).Once()
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
	mockGitServer := mocks.NewGitServer(t)

	testCases := []struct {
		name              string
		opts              *run.Options
		inputMessages     []message.Message
		expectedBreakdown string
	}{
		{
			name: "OneMessage",
			opts: &run.Options{
				GitServerClient: mockGitServer,
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
			expectedBreakdown: "[Peacock] Successfully validated 1 message(s).\n\n***\nMessage 1 will be sent to: infrastructure\n<details>\n<summary>Message Breakdown</summary>\n\nNew release of some infrastructure\nrelated things\n\n</details>\n<!-- hash: 89d156a04847b48a4e68948b83256740662f2212236fb88fa304fb28d6d6d0f6 type: breakdown -->\n",
		},
		{
			name: "MultipleMessages&MultipleTeams",
			opts: &run.Options{
				GitServerClient: mockGitServer,
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
			expectedBreakdown: "[Peacock] Successfully validated 2 message(s).\n\n***\nMessage 1 will be sent to: infrastructure\n<details>\n<summary>Message Breakdown</summary>\n\nNew release of some infrastructure\nrelated things\n\n</details>\n\n\n***\nMessage 2 will be sent to: ml\n<details>\n<summary>Message Breakdown</summary>\n\nNew release of some ml\nrelated things\n\n</details>\n<!-- hash: ea4bb9fd21b0a8eb32c437883158bd6ace2969022216a1106cbefe379ad95149 type: breakdown -->\n",
		},
	}

	mockGitServer.On("GetPRComments", mock.AnythingOfType("*context.emptyCtx"), mock.Anything, mock.Anything, 0).Return(nil, nil)

	for _, tt := range testCases {

		t.Run(tt.name, func(t *testing.T) {
			actualBreakdown, err := tt.opts.GetMessageBreakdown(context.Background(), tt.inputMessages)
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

func TestOptions_SendMessage(t *testing.T) {
	slackHandler := mocks.NewMessageHandler(t)
	webhookHandler := mocks.NewMessageHandler(t)

	testCases := []struct {
		name         string
		opts         *run.Options
		inputMessage message.Message
	}{
		{
			name: "Default",
			opts: &run.Options{
				Handlers: map[string]domain.MessageHandler{
					handlers.Slack:   slackHandler,
					handlers.Webhook: webhookHandler,
				},
				Config: &feathers.Feathers{
					Teams: []feathers.Team{
						{Name: "Infrastructure", ContactType: handlers.Slack, Addresses: []string{"#SlackAdd1", "#SlackAdd2"}},
						{Name: "AllDevs", ContactType: handlers.Slack, Addresses: []string{"#SlackAdd3", "#SlackAdd4"}},
						{Name: "Product", ContactType: handlers.Webhook, Addresses: []string{"Webhook1", "Webhook2"}},
						{Name: "Support", ContactType: handlers.Webhook, Addresses: []string{"Webhook3", "Webhook4"}},
					},
				},
			},
			inputMessage: message.Message{
				TeamNames: []string{"Infrastructure", "AllDevs", "Product", "Support"},
				Content:   "Test message content",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			slackHandler.On("Send", tc.inputMessage.Content, "", []string{"#SlackAdd1", "#SlackAdd2", "#SlackAdd3", "#SlackAdd4"}).Return(nil)
			webhookHandler.On("Send", tc.inputMessage.Content, "", []string{"Webhook1", "Webhook2", "Webhook3", "Webhook4"}).Return(nil)

			err := tc.opts.SendMessage(tc.inputMessage)
			assert.NoError(t, err)
		})
	}
}
