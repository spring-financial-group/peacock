package markdown

import (
	"github.com/microcosm-cc/bluemonday"
	md "gitlab.com/golang-commonmark/markdown"
	"regexp"
	"strings"
)

// ConvertToSlack converts the basic Markdown syntax into Slack Markup.
func ConvertToSlack(markdown string) string {
	// Remove carriage returns
	markdown = strings.ReplaceAll(markdown, "\r\n", "\n")

	var regex *regexp.Regexp
	// Convert bullets (* -> •)
	regex = regexp.MustCompile(`(^|\n)(|\s+)\*\s`)
	markdown = regex.ReplaceAllStringFunc(markdown, func(s string) string {
		return strings.ReplaceAll(s, "*", "•")
	})

	// Convert bullets (- -> •)
	regex = regexp.MustCompile(`(^|\n)(|\s+)-\s`)
	markdown = regex.ReplaceAllStringFunc(markdown, func(s string) string {
		return strings.ReplaceAll(s, "-", "•")
	})

	// Convert headings to bold (## -> *)
	regex = regexp.MustCompile(`(?m)((^\t? {0,15}#{1,4} +)(.+))`)
	markdown = regex.ReplaceAllStringFunc(markdown, func(s string) string {
		// In case someone decides to use a heading with emboldening we should strip the **
		r := regexp.MustCompile(`(?miU)((\*\*)(.+)(\*\*))`)
		return r.ReplaceAllString(s, "$3")
	})
	// Then we can remove the headers
	markdown = regex.ReplaceAllString(markdown, "*$3*")

	// Convert bold (** -> *)
	regex = regexp.MustCompile(`(?miU)((\*\*).+(\*\*))`)
	markdown = regex.ReplaceAllStringFunc(markdown, func(s string) string {
		return strings.ReplaceAll(s, "**", "*")
	})

	// Convert bold (__ -> *)
	regex = regexp.MustCompile(`(?miU)((__).+(__))`)
	markdown = regex.ReplaceAllStringFunc(markdown, func(s string) string {
		return strings.ReplaceAll(s, "__", "*")
	})

	// Convert URLs ([text](url) -> <url|text>)
	regex = regexp.MustCompile(`\[([^]]+)]\(([^)]+)\)`)
	markdown = regex.ReplaceAllString(markdown, "<$2|$1>")
	return markdown
}

// ConvertToHTML converts the Markdown syntax into HTML and sanitises the result.
func ConvertToHTML(markdown string) string {
	mdParser := md.New(md.HTML(true))
	unsafeHTML := mdParser.RenderToString([]byte(markdown))
	safeHTML := bluemonday.UGCPolicy().Sanitize(unsafeHTML)

	// the parser converts headers to <h1> tags, but we want <header> tags to make the
	// notifications consistent
	regex := regexp.MustCompile(`(?miU)((<h\d>)(.+)(</h\d>))`)
	return regex.ReplaceAllString(safeHTML, "<header>$3</header>")
}
