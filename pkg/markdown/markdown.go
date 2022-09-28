package markdown

import (
	"regexp"
	"strings"
)

// ConvertToSlack converts the basic Markdown syntax into Slack Markup.
func ConvertToSlack(markdown string) string {
	// Remove carriage returns
	markdown = strings.ReplaceAll(markdown, "\r\n", "\n")

	var regex *regexp.Regexp
	// Convert bullets
	regex = regexp.MustCompile(`(^|\n)\*\s`)
	markdown = regex.ReplaceAllString(markdown, "$1• ")

	// Convert bold (**)
	regex = regexp.MustCompile(`(?miU)((\*\*).+(\*\*))`)
	markdown = regex.ReplaceAllStringFunc(markdown, func(s string) string {
		return strings.ReplaceAll(s, "**", "*")
	})

	// Convert bold (__)
	regex = regexp.MustCompile(`(?miU)((__).+(__))`)
	markdown = regex.ReplaceAllStringFunc(markdown, func(s string) string {
		return strings.ReplaceAll(s, "__", "*")
	})

	// Convert headings to bold
	regex = regexp.MustCompile(`(?m)((^\t? {0,15}#{1,4} +)(.+))`)
	markdown = regex.ReplaceAllString(markdown, "*$3*")
	return markdown
}
