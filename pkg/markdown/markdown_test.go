package markdown_test

import (
	"github.com/spring-financial-group/peacock/pkg/markdown"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMarkdown_Converters(t *testing.T) {
	testCases := []struct {
		name          string
		inputMarkdown string
		expectedSlack string
		expectedHTML  string
	}{
		{
			name:          "HeaderAndEmbolden",
			inputMarkdown: "### **Promoted Services**",
			expectedSlack: "*Promoted Services*",
			expectedHTML:  "<header><strong>Promoted Services</strong></header>\n",
		},
		{
			name:          "CarriageReturn",
			inputMarkdown: "First Sentence\r\nSecond Sentence\r\n",
			expectedSlack: "First Sentence\nSecond Sentence\n",
			expectedHTML:  "<p>First Sentence\nSecond Sentence</p>\n",
		},
		{
			name:          "HeadingReplacement",
			inputMarkdown: "# Heading\n## Subheading\n",
			expectedSlack: "*Heading*\n*Subheading*\n",
			expectedHTML:  "<header>Heading</header>\n<header>Subheading</header>\n",
		},
		{
			name:          "BulletReplacement(*)",
			inputMarkdown: "* Bullet One\n* Bullet2",
			expectedSlack: "• Bullet One\n• Bullet2",
			expectedHTML:  "<ul>\n<li>Bullet One</li>\n<li>Bullet2</li>\n</ul>\n",
		},
		{
			name:          "BulletReplacement(-)",
			inputMarkdown: "- Bullet One\n- Bullet2",
			expectedSlack: "• Bullet One\n• Bullet2",
			expectedHTML:  "<ul>\n<li>Bullet One</li>\n<li>Bullet2</li>\n</ul>\n",
		},
		{
			name:          "BoldReplacement(**)",
			inputMarkdown: "**Bold Title**\n**Other Bold Title**",
			expectedSlack: "*Bold Title*\n*Other Bold Title*",
			expectedHTML:  "<p><strong>Bold Title</strong>\n<strong>Other Bold Title</strong></p>\n",
		},
		{
			name:          "BoldReplacement(__)",
			inputMarkdown: "__Bold Title__\n__Other Bold Title__",
			expectedSlack: "*Bold Title*\n*Other Bold Title*",
			expectedHTML:  "<p><strong>Bold Title</strong>\n<strong>Other Bold Title</strong></p>\n",
		},
		{
			name:          "URLReplacement",
			inputMarkdown: "[Some Text](https://github.com/spring-financial-group/peacock)",
			expectedSlack: "<https://github.com/spring-financial-group/peacock|Some Text>",
			expectedHTML:  "<p><a href=\"https://github.com/spring-financial-group/peacock\" rel=\"nofollow\">Some Text</a></p>\n",
		},
		{
			name:          "TestPRTemplate",
			inputMarkdown: "### **Promoted Services**\n_Which services are being promoted?_\n* Peacock \n\n### **What functionality is being released?**\n_What features/bug fixes are present?_\n\n* All the features\n* All the bugs\n",
			expectedSlack: "*Promoted Services*\n_Which services are being promoted?_\n• Peacock \n\n*What functionality is being released?*\n_What features/bug fixes are present?_\n\n• All the features\n• All the bugs\n",
			expectedHTML:  "<header><strong>Promoted Services</strong></header>\n<p><em>Which services are being promoted?</em></p>\n<ul>\n<li>Peacock</li>\n</ul>\n<header><strong>What functionality is being released?</strong></header>\n<p><em>What features/bug fixes are present?</em></p>\n<ul>\n<li>All the features</li>\n<li>All the bugs</li>\n</ul>\n",
		},
		{
			name:          "NestedBulletReplacement(*)",
			inputMarkdown: "* New queries added to Product/Summary endpoint:\n   * Return products by product class\n   * Return products that were available for a given date",
			expectedSlack: "• New queries added to Product/Summary endpoint:\n   • Return products by product class\n   • Return products that were available for a given date",
			expectedHTML:  "<ul>\n<li>New queries added to Product/Summary endpoint:\n<ul>\n<li>Return products by product class</li>\n<li>Return products that were available for a given date</li>\n</ul>\n</li>\n</ul>\n",
		},
		{
			name:          "NestedBulletReplacement(-)",
			inputMarkdown: "- New queries added to Product/Summary endpoint:\n   - Return products by product class\n   - Return products that were available for a given date",
			expectedSlack: "• New queries added to Product/Summary endpoint:\n   • Return products by product class\n   • Return products that were available for a given date",
			expectedHTML:  "<ul>\n<li>New queries added to Product/Summary endpoint:\n<ul>\n<li>Return products by product class</li>\n<li>Return products that were available for a given date</li>\n</ul>\n</li>\n</ul>\n",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			actualSlack := markdown.ConvertToSlack(tt.inputMarkdown)
			actualHTML := markdown.ConvertToHTML(tt.inputMarkdown)
			assert.Equal(t, tt.expectedSlack, actualSlack, "Slack Conversion")
			assert.Equal(t, tt.expectedHTML, actualHTML, "HTML Conversion")
		})
	}
}
