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
			inputMarkdown: "# Peacock\n## Message\n### Notify infrastructure\nTest Content",
			expectedMessages: []message.Message{
				{
					TeamNames: []string{"infrastructure"},
					Content:   "Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:          "HeadingsInContent",
			inputMarkdown: "# Peacock\n## Message\n### Notify infrastructure\n### Test Content\nThis is some content with headers\n#### Another different header",
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
			inputMarkdown: "# Peacock\n# Peacock Release Format\n***\n## Message\n### Notify infrastructure\nTest Content",
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
			shouldError:      true,
		},
		{
			name:             "MoreTeamsThanMessages",
			inputMarkdown:    "# Peacock\n## Message\n### Notify infrastructure\n### Notify ml\nTest Content",
			expectedMessages: nil,
			shouldError:      true,
		},
		{
			name:             "NoMessages",
			inputMarkdown:    "# Peacock\n### Notify Team\n",
			expectedMessages: nil,
			shouldError:      true,
		},
		{
			name:             "NoTeams",
			inputMarkdown:    "# Peacock\n## Message\nTest Content",
			expectedMessages: nil,
			shouldError:      true,
		},
		{
			name:          "MultipleMessages",
			inputMarkdown: "# Peacock\n## Message\n### Notify infrastructure\nTest Content\n## Message\n### Notify ML\nMore test content",
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
			inputMarkdown: "# Peacock\n## Message\n### Notify infrastructure, ml, allDevs\nTest Content\n",
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
			inputMarkdown: "\n\n\n# Peacock\n## Message\n### Notify infrastructure\nTest Content\n\n\n",
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
			inputMarkdown: "# Peacock\n## Message\n### Notify infrastructure\nThis is an example\nThat runs\nAcross multiple\nlines",
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
			inputMarkdown: "# Peacock\n## Message\n### Notify infrastructure\nHere's a list of what we've done\n\t- Fixes\n\t- Features\n\t- bugs",
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
			inputMarkdown: "# Peacock\n## Message\n### Notify infrastructure   \nTest Content",
			expectedMessages: []message.Message{
				{
					TeamNames: []string{"infrastructure"},
					Content:   "Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:             "NoWhitespaceBeforeTeamName",
			inputMarkdown:    "# Peacock\n## Message\n### Notifyinfrastructure\nTest Content",
			expectedMessages: nil,
			shouldError:      true,
		},
		{
			name:          "WhitespaceAfterMessageHeader",
			inputMarkdown: "# Peacock\n## Message  \n### Notify infrastructure\nTest Content",
			expectedMessages: []message.Message{
				{
					TeamNames: []string{"infrastructure"},
					Content:   "Test Content",
				},
			},
			shouldError: false,
		},
		{
			name:             "NoWhitespaceBeforeMessage",
			inputMarkdown:    "# Peacock\n##Message\n### Notify infrastructure\nTest Content",
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
