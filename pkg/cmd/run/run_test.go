package run_test

import (
	"fmt"
	"github.com/google/go-github/v47/github"
	"github.com/spring-financial-group/peacock/pkg/cmd/run"
	"github.com/spring-financial-group/peacock/pkg/config"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/domain/mocks"
	"github.com/spring-financial-group/peacock/pkg/handlers"
	"github.com/spring-financial-group/peacock/pkg/message"
	"github.com/spring-financial-group/peacock/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestOptions_Run(t *testing.T) {
	mockGit := mocks.NewGit(t)
	mockSlackHander := mocks.NewMessageHandler(t)

	testCases := []struct {
		name        string
		opts        *run.Options
		pr          *github.PullRequest
		shouldError bool
	}{
		{
			name: "NonDryRun",
			opts: &run.Options{
				PRNumber:     1,
				GitServerURL: "https://github.com",
				GitToken:     "testGitToken",
				DryRun:       false,
				SlackToken:   "testSlackToken",
				Git:          mockGit,
				Handlers:     map[string]domain.MessageHandler{handlers.Slack: mockSlackHander},
				Config: &config.Config{
					Teams: []config.Team{
						{
							Name:        "infrastructure",
							ContactType: handlers.Slack,
							Addresses:   []string{"TestAdd", "TestAdd2"},
						},
					},
				},
			},
			pr: &github.PullRequest{
				Body: utils.NewPtr(
					"## Message\n### Team: infrastructure\nTest Content",
				),
			},
		},
		{
			name: "DryRun",
			opts: &run.Options{
				PRNumber:     1,
				GitServerURL: "https://github.com",
				GitToken:     "testGitToken",
				DryRun:       true,
				SlackToken:   "testSlackToken",
				Git:          mockGit,
				Handlers:     map[string]domain.MessageHandler{handlers.Slack: mockSlackHander},
				Config: &config.Config{
					Teams: []config.Team{
						{
							Name:        "infrastructure",
							ContactType: handlers.Slack,
							Addresses:   []string{"TestAdd", "TestAdd2"},
						},
					},
				},
			},
			pr: &github.PullRequest{
				Body: utils.NewPtr(
					"## Message\n### Team: infrastructure\nTest Content",
				),
			},
		},
	}

	for _, tt := range testCases {
		mockGit.On("GetPullRequest", mock.AnythingOfType("*context.emptyCtx"), tt.opts.PRNumber).Return(tt.pr, nil).Once()
		if tt.opts.DryRun {
			mockGit.On("CommentOnPR", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("*github.PullRequest"), mock.AnythingOfType("string")).Return(nil).Once()
		}

		for _, team := range tt.opts.Config.Teams {
			if !tt.opts.DryRun {
				mockSlackHander.On("Send", "Test Content", team.Addresses).Return(nil).Once()
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
				Config: &config.Config{
					Teams: []config.Team{
						{Name: "infrastructure"},
					},
				},
			},
			inputMessages: []message.Message{
				{
					TeamName: "infrastructure",
					Content:  "New release of some infrastructure\nrelated things",
				},
			},
			expectedBreakdown: "### Validation\nSuccessfully parsed 1 message(s)\n1/1 teams in config to notify\n***\n### Message [1/1]\n#### Team: infrastructure\n#### Contact Type: \n#### Content:\nNew release of some infrastructure\nrelated things",
		},
		{
			name: "MultipleMessages&MultipleTeams",
			opts: &run.Options{
				Config: &config.Config{
					Teams: []config.Team{
						{Name: "infrastructure"},
						{Name: "ml"},
					},
				},
			},
			inputMessages: []message.Message{
				{
					TeamName: "infrastructure",
					Content:  "New release of some infrastructure\nrelated things",
				},
				{
					TeamName: "ml",
					Content:  "New release of some ml\nrelated things",
				},
			},
			expectedBreakdown: "### Validation\nSuccessfully parsed 2 message(s)\n2/2 teams in config to notify\n***\n### Message [1/2]\n#### Team: infrastructure\n#### Contact Type: \n#### Content:\nNew release of some infrastructure\nrelated things\n***\n### Message [2/2]\n#### Team: ml\n#### Contact Type: \n#### Content:\nNew release of some ml\nrelated things",
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
				Config: &config.Config{
					Teams: []config.Team{
						{Name: "infrastructure", ContactType: handlers.Slack},
					},
				},
			},
			inputMessages: []message.Message{
				{
					TeamName: "infrastructure",
					Content:  "some content",
				},
			},
			shouldError: false,
		},
		{
			name: "TeamDoesNotExist",
			opts: &run.Options{
				Handlers: map[string]domain.MessageHandler{handlers.Slack: mocks.NewMessageHandler(t)},
				Config: &config.Config{
					Teams: []config.Team{
						{Name: "infrastructure", ContactType: handlers.Slack},
					},
				},
			},
			inputMessages: []message.Message{
				{
					TeamName: "ml",
					Content:  "some content",
				},
			},
			shouldError: true,
		},
		{
			name: "HandlerDoesNotExist",
			opts: &run.Options{
				Handlers: map[string]domain.MessageHandler{},
				Config: &config.Config{
					Teams: []config.Team{
						{Name: "infrastructure", ContactType: handlers.Slack},
					},
				},
			},
			inputMessages: []message.Message{
				{
					TeamName: "infrastructure",
					Content:  "some content",
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