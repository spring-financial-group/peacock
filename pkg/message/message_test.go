package message_test

import (
	"fmt"
	"github.com/spring-financial-group/peacock/pkg/message"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		name             string
		inputMarkdown    string
		expectedMessages []message.Message
		shouldError      bool
	}{
		{
			name:          "Passing",
			inputMarkdown: "### Notify infrastructure, devs\nTest Content\n### Notify ml\nMore Test Content",
			expectedMessages: []message.Message{
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
			expectedMessages: []message.Message{
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
			expectedMessages: []message.Message{
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
			expectedMessages: []message.Message{
				{
					TeamNames: []string{"infrastructure"},
					Content:   "Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:             "NoInputMarkdown",
			inputMarkdown:    "",
			expectedMessages: nil,
			shouldError:      false,
		},
		{
			name:             "NoMessages",
			inputMarkdown:    "# Title to the PR\nSome information about the pr\n",
			expectedMessages: nil,
			shouldError:      false,
		},
		{
			name:             "NoTeams",
			inputMarkdown:    "### Notify ",
			expectedMessages: nil,
			shouldError:      false,
		},
		{
			name:          "MultipleMessages",
			inputMarkdown: "### Notify infrastructure\nTest Content\n### Notify ML\nMore test content\n",
			expectedMessages: []message.Message{
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
			expectedMessages: []message.Message{
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
			expectedMessages: []message.Message{
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
			expectedMessages: []message.Message{
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
			expectedMessages: []message.Message{
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
			expectedMessages: []message.Message{
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
			expectedMessages: []message.Message{
				{
					TeamNames: []string{"infrastructure", "ml", "product"},
					Content:   "Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:          "NoWhitespaceBeforeTeamName",
			inputMarkdown: "# Peacock\r\n## Message\n### Notifyinfrastructure\nTest Content",
			expectedMessages: []message.Message{
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
			actualMessages, err := message.ParseMessagesFromMarkdown(tt.inputMarkdown)
			if tt.shouldError {
				fmt.Println("expected error: " + err.Error())
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedMessages, actualMessages)
		})
	}
}
