package diffs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_jsonHandler_Diff(t *testing.T) {
	t.Run("nested paths with additions, changes and removals", func(t *testing.T) {
		base := []byte(`{"a": 1, "obj": {"x": "old"}}`)
		head := []byte(`{"a": 1, "obj": {"x": "new", "y": true}, "arr": [1, 2]}`)

		sections, err := (jsonHandler{}).Diff("config.json", base, head)
		assert.NoError(t, err)
		assert.Len(t, sections, 1)

		s := sections[0]
		assert.Equal(t, "<code>config.json</code> — 4 paths changed", s.Summary)
		assert.Contains(t, s.Body, "| `arr[0]` | ➕ added | 1 |")
		assert.Contains(t, s.Body, "| `arr[1]` | ➕ added | 2 |")
		assert.Contains(t, s.Body, "| `obj.x` | ✏ changed | old → new |")
		assert.Contains(t, s.Body, "| `obj.y` | ➕ added | true |")
	})

	t.Run("removed paths have no details", func(t *testing.T) {
		base := []byte(`{"gone": "value", "kept": 1}`)
		head := []byte(`{"kept": 1}`)

		sections, err := (jsonHandler{}).Diff("config.json", base, head)
		assert.NoError(t, err)
		assert.Len(t, sections, 1)
		assert.Contains(t, sections[0].Body, "| `gone` | ➖ removed |  |")
	})

	t.Run("identical documents return no sections", func(t *testing.T) {
		doc := []byte(`{"a": [1, 2], "b": {"c": "d"}}`)
		sections, err := (jsonHandler{}).Diff("config.json", doc, doc)
		assert.NoError(t, err)
		assert.Empty(t, sections)
	})

	t.Run("new document returns a count summary", func(t *testing.T) {
		head := []byte(`{"a": 1, "b": 2}`)
		sections, err := (jsonHandler{}).Diff("config.json", nil, head)
		assert.NoError(t, err)
		assert.Len(t, sections, 1)
		assert.Equal(t, "<code>config.json</code> — 2 paths (new)", sections[0].Summary)
		assert.Empty(t, sections[0].Body)
	})

	t.Run("deleted document returns a removed summary", func(t *testing.T) {
		base := []byte(`{"a": 1}`)
		sections, err := (jsonHandler{}).Diff("config.json", base, nil)
		assert.NoError(t, err)
		assert.Len(t, sections, 1)
		assert.Equal(t, "<code>config.json</code> — removed", sections[0].Summary)
	})

	t.Run("invalid json errors", func(t *testing.T) {
		_, err := (jsonHandler{}).Diff("config.json", []byte("not json"), []byte(`{}`))
		assert.Error(t, err)
	})
}

func Test_truncate(t *testing.T) {
	t.Run("short values are untouched", func(t *testing.T) {
		assert.Equal(t, "short", truncate("short"))
	})

	t.Run("long values are truncated to the max length", func(t *testing.T) {
		long := strings.Repeat("a", maxValueLen+50)
		truncated := truncate(long)
		assert.True(t, strings.HasPrefix(truncated, strings.Repeat("a", maxValueLen-1)))
		assert.True(t, strings.HasSuffix(truncated, "…"))
	})
}
