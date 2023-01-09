package comment

import (
	"fmt"
	"regexp"
)

const (
	BreakdownCommentType = "breakdown"
)

var (
	re = regexp.MustCompile(`(?m)<!-- hash: ([a-zA-Z0-9]+) type: ([a-zA-Z0-9]+) -->`)
)

// GetMetadataFromComment returns the hash from a comment
func GetMetadataFromComment(comment string) (hash string, commentType string) {
	matches := re.FindStringSubmatch(comment)
	if len(matches) != 3 {
		return "", ""
	}
	return matches[1], matches[2]
}

// AddMetadataToComment adds the hash and comment type to a comment
func AddMetadataToComment(comment, hash, commentType string) string {
	return fmt.Sprintf("%s\n<!-- hash: %s type: %s -->\n", comment, hash, commentType)
}
