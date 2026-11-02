package diffs

import (
	"fmt"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/spring-financial-group/peacock/pkg/models"
	"path"
	"strings"
)

const maxDiffLines = 400

type textHandler struct{}

func (textHandler) Diff(key string, base, head []byte) ([]models.DiffSection, error) {
	if base == nil && head == nil {
		return nil, nil
	}
	name := path.Base(key)

	var diff difflib.UnifiedDiff
	summary := fmt.Sprintf("<code>%s</code> — changed", escapeCell(name))
	switch {
	case base == nil:
		summary = fmt.Sprintf("<code>%s</code> — new file", escapeCell(name))
		diff = difflib.UnifiedDiff{
			FromFile: "/dev/null",
			ToFile:   name,
			A:        nil,
			B:        difflib.SplitLines(string(head)),
			Context:  3,
		}
	case head == nil:
		summary = fmt.Sprintf("<code>%s</code> — removed", escapeCell(name))
		diff = difflib.UnifiedDiff{
			FromFile: name,
			ToFile:   "/dev/null",
			A:        difflib.SplitLines(string(base)),
			B:        nil,
			Context:  3,
		}
	default:
		if string(base) == string(head) {
			return nil, nil
		}
		diff = difflib.UnifiedDiff{
			FromFile: name,
			ToFile:   name,
			A:        difflib.SplitLines(string(base)),
			B:        difflib.SplitLines(string(head)),
			Context:  3,
		}
	}

	out, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		return nil, err
	}
	body := fmt.Sprintf("```diff\n%s\n```", capLines(out, maxDiffLines))
	return []models.DiffSection{{
		Key:     name,
		Summary: summary,
		Body:    body,
		Open:    true,
	}}, nil
}

// capLines truncates very large diffs rather than blowing the comment size budget
func capLines(s string, max int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) <= max {
		return strings.Join(lines, "\n")
	}
	return strings.Join(lines[:max], "\n") + "\n… diff truncated"
}
