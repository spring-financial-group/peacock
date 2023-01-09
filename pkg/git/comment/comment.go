package comment

import "regexp"

// GetHashFromComment returns the hash from a comment
func GetHashFromComment(comment string) string {
	re := regexp.MustCompile(`(?m)<!-- hash: ([a-zA-Z0-9]+) -->`)
	matches := re.FindStringSubmatch(comment)
	if len(matches) != 2 {
		return ""
	}
	return matches[1]
}
