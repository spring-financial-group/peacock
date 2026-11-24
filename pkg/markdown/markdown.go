package markdown

import (
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	md "gitlab.com/golang-commonmark/markdown"
)

// ConvertToSlack converts the basic Markdown syntax into Slack Markup.
func ConvertToSlack(markdown string) string {
	// Remove carriage returns
	markdown = strings.ReplaceAll(markdown, "\r\n", "\n")

	markdown = stripDetailsTags(markdown)

	var regex *regexp.Regexp
	// Convert bullets (* -> •)
	regex = regexp.MustCompile(`(^|\n)(|\s+)\*\s`)
	markdown = regex.ReplaceAllStringFunc(markdown, func(s string) string {
		return strings.ReplaceAll(s, "*", "•")
	})

	// Convert unchecked task box (- [ ] -> ☐)
	regex = regexp.MustCompile(`(^|\n)(|\s+)- \[ ] `)
	markdown = regex.ReplaceAllStringFunc(markdown, func(s string) string {
		return strings.ReplaceAll(s, "- [ ] ", "☐ ")
	})

	// Convert unchecked task box (- [x] -> ☒)
	regex = regexp.MustCompile(`-\s\[[xX]]`)
	markdown = regex.ReplaceAllString(markdown, "☒")

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

	// Convert GitHub links (ORG/REPO_NAME#PR -> <OG_TEXT|https://github.com/ORG/REPO_NAME/pull/PR>)
	regex = regexp.MustCompile(`([^/]+)/([^/]+)#(\d+)`)
	markdown = regex.ReplaceAllString(markdown, "<$0|https://github.com/$1/$2/pull/$3>")

	return markdown
}

// ConvertToHTML converts the Markdown syntax into HTML and sanitises the result.
func ConvertToHTML(markdown string) string {
	markdown = stripDetailsTags(markdown)

	mdParser := md.New(md.HTML(true), md.Breaks(true), md.Tables(true))
	unsafeHTML := mdParser.RenderToString([]byte(markdown))
	policy := bluemonday.UGCPolicy()
	policy.AllowAttrs("style").Matching(regexp.MustCompile(`^color: #[0-9a-fA-F]{6}; background-color: #[0-9a-fA-F]{6};?$`)).OnElements("span")
	safeHTML := policy.Sanitize(unsafeHTML)
	safeHTML = styleEmailTables(safeHTML)
	safeHTML = styleDiffLines(safeHTML)

	// the parser converts headers to <h1> tags, but we want <header> tags to make the
	// notifications consistent
	regex := regexp.MustCompile(`(?miU)((<h\d>)(.+)(</h\d>))`)
	return regex.ReplaceAllString(safeHTML, "<header>$3</header>")
}

func styleEmailTables(html string) string {
	tableRegex := regexp.MustCompile(`(?s)<table>(.*?)</table>`)
	return tableRegex.ReplaceAllStringFunc(html, func(table string) string {
		table = strings.ReplaceAll(table, `<table>`, `<table style="border-collapse: collapse; width: 100%; margin: 12px 0;">`)
		return regexp.MustCompile(`<(th|td)>`).ReplaceAllString(table, `<$1 style="border: 1px solid #d0d7de; padding: 8px 12px; text-align: left;">`)
	})
}

func styleDiffLines(html string) string {
	codeBlockRegex := regexp.MustCompile(`(?s)(<pre><code(?: class="language-[^"]*")?>)(.*?)(</code></pre>)`)
	return codeBlockRegex.ReplaceAllStringFunc(html, func(block string) string {
		parts := codeBlockRegex.FindStringSubmatch(block)
		lines := strings.Split(parts[2], "\n")
		for i, line := range lines {
			switch {
			case strings.HasPrefix(line, "+"):
				lines[i] = `<span style="color: #1a7f37; background-color: #dafbe1;">` + line + `</span>`
			case strings.HasPrefix(line, "-"):
				lines[i] = `<span style="color: #cf222e; background-color: #ffebe9;">` + line + `</span>`
			}
		}
		opening := strings.Replace(parts[1], `<pre>`, `<pre style="border: 1px solid #d0d7de; background-color: #f6f8fa; padding: 12px; white-space: pre-wrap;">`, 1)
		return opening + strings.Join(lines, "\n") + parts[3]
	})
}

// stripDetailsTags handles GitHub-style <details>/<summary> markup so it renders
// cleanly: <summary>X</summary> is rewritten to a heading, and the surrounding
// <details> tags are removed.
// Each tag also consumes up to one adjacent blank line on its inner side
// (after the opening/summary tags, before the closing tag) so the blank line
// GitHub requires for the inner content to parse as markdown doesn't leak
// into the output. Any further blank lines the author added are preserved.
func stripDetailsTags(markdown string) string {
	summaryRegex := regexp.MustCompile(`(?i)<summary(?:\s[^>]*)?>(.*?)</summary>\n{0,2}`)
	markdown = summaryRegex.ReplaceAllString(markdown, "## $1\n")

	openRegex := regexp.MustCompile(`(?i)<details(?:\s[^>]*)?>\n{0,2}`)
	markdown = openRegex.ReplaceAllString(markdown, "")

	closeRegex := regexp.MustCompile(`(?i)\n{0,2}</details>`)
	return closeRegex.ReplaceAllString(markdown, "")
}
