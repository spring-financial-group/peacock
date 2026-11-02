package diffs

import (
	"fmt"
	"testing"

	"github.com/spring-financial-group/peacock/pkg/models"
	"github.com/stretchr/testify/assert"
)

func Test_detectInnerFormat(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected models.DiffFormat
	}{
		{
			name:     "JSONSchema",
			input:    `{"$schema": "https://json-schema.org/draft/2020-12/schema", "properties": {}}`,
			expected: models.DiffFormatJSONSchema,
		},
		{
			name:     "JSONWithoutSchemaKey",
			input:    `{"some": "json"}`,
			expected: models.DiffFormatJSON,
		},
		{
			name:     "JSONArray",
			input:    `[1, 2, 3]`,
			expected: models.DiffFormatJSON,
		},
		{
			name:     "EmptySchemaKey",
			input:    `{"$schema": ""}`,
			expected: models.DiffFormatJSON,
		},
		{
			name:     "NotJSON",
			input:    "just some text",
			expected: models.DiffFormatText,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, detectInnerFormat([]byte(tt.input)))
		})
	}
}

func Test_renderTable(t *testing.T) {
	t.Run("rows under the cap are all rendered", func(t *testing.T) {
		rows := []row{{path: "a", change: addedMarker, details: "string"}}
		table, truncated := renderTable(rows)
		assert.Equal(t, "| Field | Change | Details |\n|---|---|---|\n| `a` | ➕ added | string |\n", table)
		assert.Equal(t, 0, truncated)
	})

	t.Run("rows over the cap are truncated", func(t *testing.T) {
		var rows []row
		for i := 0; i < maxTableRows+5; i++ {
			rows = append(rows, row{path: fmt.Sprintf("field%d", i), change: addedMarker, details: "string"})
		}
		table, truncated := renderTable(rows)
		assert.Equal(t, 5, truncated)
		assert.Equal(t, maxTableRows+2, countLines(table))
	})

	t.Run("description column appears only when a row carries one", func(t *testing.T) {
		withDesc := []row{{path: "a", change: addedMarker, details: "string", description: "The a value"}}
		table, _ := renderTable(withDesc)
		assert.Contains(t, table, "| Field | Change | Details | Description |\n|---|---|---|---|\n")
		assert.Contains(t, table, "| `a` | ➕ added | string | The a value |")

		mixed := []row{
			{path: "a", change: addedMarker, details: "string", description: "The a value"},
			{path: "b", change: addedMarker, details: "integer"},
		}
		table, _ = renderTable(mixed)
		assert.Contains(t, table, "| `b` | ➕ added | integer |  |")
	})

	t.Run("cells are escaped", func(t *testing.T) {
		rows := []row{{path: "a|b", change: addedMarker, details: "line1\nline2"}}
		table, _ := renderTable(rows)
		assert.Contains(t, table, "| `a\\|b` | ➕ added | line1 line2 |")
	})
}

func Test_escapeCell(t *testing.T) {
	assert.Equal(t, `a\|b`, escapeCell("a|b"))
	assert.Equal(t, "a b", escapeCell("a\nb"))
	assert.Equal(t, "a", escapeCell("  a  "))
}

func countLines(s string) int {
	n := 0
	for _, c := range s {
		if c == '\n' {
			n++
		}
	}
	return n
}
