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
			inputMarkdown: "## Message\n### Team: infrastructure\nTest Content",
			expectedMessages: []message.Message{
				{
					TeamName: "infrastructure",
					Content:  "Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:          "HeadingsInContent",
			inputMarkdown: "## Message\n### Team: infrastructure\n### Test Content\nThis is some content with headers\n#### Another different header",
			expectedMessages: []message.Message{
				{
					TeamName: "infrastructure",
					Content:  "### Test Content\nThis is some content with headers\n#### Another different header",
				},
			},
			shouldError: false,
		},
		{
			name:          "PrefaceToMessages",
			inputMarkdown: "# Peacock Release Format\n***\n## Message\n### Team: infrastructure\nTest Content",
			expectedMessages: []message.Message{
				{
					TeamName: "infrastructure",
					Content:  "Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:             "NoInputMarkdown",
			inputMarkdown:    "",
			expectedMessages: nil,
			shouldError:      true,
		},
		{
			name:             "MoreTeamsThanMessages",
			inputMarkdown:    "## Message\n### Team: infrastructure\n### Team: ml\nTest Content",
			expectedMessages: nil,
			shouldError:      true,
		},
		{
			name:             "NoMessages",
			inputMarkdown:    "### Team: Team\n",
			expectedMessages: nil,
			shouldError:      true,
		},
		{
			name:             "NoTeams",
			inputMarkdown:    "## Message\nTest Content",
			expectedMessages: nil,
			shouldError:      true,
		},
		{
			name:          "MultipleMessages",
			inputMarkdown: "## Message\n### Team: infrastructure\nTest Content\n## Message\n### Team: ML\nMore test content",
			expectedMessages: []message.Message{
				{
					TeamName: "infrastructure",
					Content:  "Test Content",
				},
				{
					TeamName: "ML",
					Content:  "More test content",
				},
			},
			shouldError: false,
		},
		{
			name:          "AdditionalNewLines",
			inputMarkdown: "\n\n\n## Message\n### Team: infrastructure\nTest Content\n\n\n",
			expectedMessages: []message.Message{
				{
					TeamName: "infrastructure",
					Content:  "Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:          "MultiLineContent",
			inputMarkdown: "## Message\n### Team: infrastructure\nThis is an example\nThat runs\nAcross multiple\nlines",
			expectedMessages: []message.Message{
				{
					TeamName: "infrastructure",
					Content:  "This is an example\nThat runs\nAcross multiple\nlines",
				},
			},
			shouldError: false,
		},
		{
			name:          "Lists",
			inputMarkdown: "## Message\n### Team: infrastructure\nHere's a list of what we've done\n\t- Fixes\n\t- Features\n\t- bugs",
			expectedMessages: []message.Message{
				{
					TeamName: "infrastructure",
					Content:  "Here's a list of what we've done\n\t- Fixes\n\t- Features\n\t- bugs",
				},
			},
			shouldError: false,
		},
		{
			name:          "WhitespaceAfterTeamName",
			inputMarkdown: "## Message\n### Team: infrastructure   \nTest Content",
			expectedMessages: []message.Message{
				{
					TeamName: "infrastructure",
					Content:  "Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:             "NoWhitespaceBeforeTeamName",
			inputMarkdown:    "## Message\n### Team:infrastructure\nTest Content",
			expectedMessages: nil,
			shouldError:      true,
		},
		{
			name:          "WhitespaceAfterMessageHeader",
			inputMarkdown: "## Message  \n### Team: infrastructure\nTest Content",
			expectedMessages: []message.Message{
				{
					TeamName: "infrastructure",
					Content:  "Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:             "NoWhitespaceBeforeMessage",
			inputMarkdown:    "##Message\n### Team: infrastructure\nTest Content",
			expectedMessages: nil,
			shouldError:      true,
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
