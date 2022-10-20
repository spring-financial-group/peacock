package markdown_test

import (
	"github.com/spring-financial-group/peacock/pkg/markdown"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMarkdown_ConvertMarkdownToSlack(t *testing.T) {
	testCases := []struct {
		name          string
		inputMarkdown string
		expectedSlack string
		shouldError   bool
	}{
		{
			name:          "HeaderAndEmbolden",
			inputMarkdown: "### **Promoted Services**",
			expectedSlack: "*Promoted Services*",
		},
		{
			name:          "CarriageReturn",
			inputMarkdown: "First Sentence\r\nSecond Sentence\r\n",
			expectedSlack: "First Sentence\nSecond Sentence\n",
		},
		{
			name:          "HeadingReplacement",
			inputMarkdown: "# Heading\n## Subheading\n",
			expectedSlack: "*Heading*\n*Subheading*\n",
		},
		{
			name:          "BulletReplacement",
			inputMarkdown: "* Bullet One\n* Bullet2",
			expectedSlack: "• Bullet One\n• Bullet2",
		},
		{
			name:          "BoldReplacement(**)",
			inputMarkdown: "**Bold Title**\n**Other Bold Title**",
			expectedSlack: "*Bold Title*\n*Other Bold Title*",
		},
		{
			name:          "BoldReplacement(__)",
			inputMarkdown: "__Bold Title__\n__Other Bold Title__",
			expectedSlack: "*Bold Title*\n*Other Bold Title*",
		},
		{
			name:          "TestPRTemplate",
			inputMarkdown: "# Service Promotions\n\n**Promoted Services**\n\n_Which services are being promoted?_\n_eg._\n* Api Gateway\n* Questions Library\n\n**What functionality is being released?**\n_eg._\n* Questions Library initial release (but not connected to anything yet)\n\n**Risk Of Release**\nVery Low",
			expectedSlack: "*Service Promotions*\n\n*Promoted Services*\n\n_Which services are being promoted?_\n_eg._\n• Api Gateway\n• Questions Library\n\n*What functionality is being released?*\n_eg._\n• Questions Library initial release (but not connected to anything yet)\n\n*Risk Of Release*\nVery Low",
		},
		{
			name:          "URLReplacement",
			inputMarkdown: "[Some Text](https://github.com/spring-financial-group/peacock)",
			expectedSlack: "<https://github.com/spring-financial-group/peacock|Some Text>",
		},
		{
			name:          "Full",
			inputMarkdown: "### **Promoted Services**\n_Which services are being promoted?_\n* Peacock \n\n### **What functionality is being released?**\n_What features/bug fixes are present?_\n\n* All the features\n* All the bugs\n",
			expectedSlack: "*Promoted Services*\n_Which services are being promoted?_\n• Peacock \n\n*What functionality is being released?*\n_What features/bug fixes are present?_\n\n• All the features\n• All the bugs\n",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			actualSlack := markdown.ConvertToSlack(tt.inputMarkdown)
			assert.Equal(t, tt.expectedSlack, actualSlack)
		})
	}
}
