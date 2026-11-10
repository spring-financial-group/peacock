package markdown_test

import (
	"testing"

	"github.com/spring-financial-group/peacock/pkg/markdown"
	"github.com/stretchr/testify/assert"
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
			expectedHTML:  "<p>First Sentence<br>\nSecond Sentence</p>\n",
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
			expectedHTML:  "<p><strong>Bold Title</strong><br>\n<strong>Other Bold Title</strong></p>\n",
		},
		{
			name:          "BoldReplacement(__)",
			inputMarkdown: "__Bold Title__\n__Other Bold Title__",
			expectedSlack: "*Bold Title*\n*Other Bold Title*",
			expectedHTML:  "<p><strong>Bold Title</strong><br>\n<strong>Other Bold Title</strong></p>\n",
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
		{
			name:          "UnTickedTaskListReplacement(- [ ])",
			inputMarkdown: "- [ ] No impact to reporting\n- [ ] No impact to downstream services",
			expectedSlack: "☐ No impact to reporting\n☐ No impact to downstream services",
			expectedHTML:  "<ul>\n<li>[ ] No impact to reporting</li>\n<li>[ ] No impact to downstream services</li>\n</ul>\n",
		},
		{
			name:          "TickedTaskListReplacement(- [x])",
			inputMarkdown: "- [x] No impact to reporting\n- [X] No impact to downstream services",
			expectedSlack: "☒ No impact to reporting\n☒ No impact to downstream services",
			expectedHTML:  "<ul>\n<li>[x] No impact to reporting</li>\n<li>[X] No impact to downstream services</li>\n</ul>\n",
		},
		{
			name:          "GithubLinkReplacement",
			inputMarkdown: "spring-financial-group/mqube-property-service#770",
			expectedSlack: "<spring-financial-group/mqube-property-service#770|https://github.com/spring-financial-group/mqube-property-service/pull/770>",
			expectedHTML:  "<p>spring-financial-group/mqube-property-service#770</p>\n",
		},
		{
			name:          "DetailsBlockWithSummary",
			inputMarkdown: "<details>\n<summary>Click to expand</summary>\n\nHidden content here\n\n</details>",
			expectedSlack: "*Click to expand*\nHidden content here",
			expectedHTML:  "<header>Click to expand</header>\n<p>Hidden content here</p>\n",
		},
		{
			name:          "DetailsBlockWithOpenAttribute",
			inputMarkdown: "<details open>\n<summary>Details</summary>\n\nMore info\n\n</details>",
			expectedSlack: "*Details*\nMore info",
			expectedHTML:  "<header>Details</header>\n<p>More info</p>\n",
		},
		{
			name:          "DetailsBlockWithoutSummary",
			inputMarkdown: "<details>\nJust some collapsible text\n</details>",
			expectedSlack: "Just some collapsible text",
			expectedHTML:  "<p>Just some collapsible text</p>\n",
		},
		{
			name:          "DetailsBlockCollapsesSurroundingBlankLines",
			inputMarkdown: "<details open>\n<summary>Some Important Summary</summary>\n\n### Important Header\n**A Service**\n- A feature\n\n</details>",
			expectedSlack: "*Some Important Summary*\n*Important Header*\n*A Service*\n• A feature",
			expectedHTML:  "<header>Some Important Summary</header>\n<header>Important Header</header>\n<p><strong>A Service</strong></p>\n<ul>\n<li>A feature</li>\n</ul>\n",
		},
		{
			name:          "DetailsBlockPreservesIntentionalBlankLines",
			inputMarkdown: "<details>\n<summary>Title</summary>\n\n\nFirst paragraph\n\n\n</details>",
			expectedSlack: "*Title*\n\nFirst paragraph\n",
			expectedHTML:  "<header>Title</header>\n<p>First paragraph</p>\n",
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
