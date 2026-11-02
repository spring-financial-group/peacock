package diffs

import (
	"encoding/json"
	"fmt"
	"github.com/spring-financial-group/peacock/pkg/models"
	"strings"
)

// Handler renders the diff of one matched file between its base and head versions.
type Handler interface {
	// Diff returns one section per sub-document (e.g. each ConfigMap data key),
	// or none when the file has no effective changes.
	// A nil side means the document is new (no base) or deleted (no head).
	Diff(key string, base, head []byte) ([]models.DiffSection, error)
}

func getHandler(format models.DiffFormat) (Handler, error) {
	switch format {
	case models.DiffFormatText:
		return textHandler{}, nil
	case models.DiffFormatJSON:
		return jsonHandler{}, nil
	case models.DiffFormatJSONSchema:
		return jsonSchemaHandler{}, nil
	case models.DiffFormatConfigMap:
		return configmapHandler{}, nil
	}
	return nil, fmt.Errorf("unsupported diff format %s", format)
}

// detectInnerFormat sniffs a document's content to pick a renderer for it
func detectInnerFormat(b []byte) models.DiffFormat {
	var doc any
	if json.Unmarshal(b, &doc) != nil {
		return models.DiffFormatText
	}
	if m, ok := doc.(map[string]any); ok {
		if s, ok := m["$schema"].(string); ok && s != "" {
			return models.DiffFormatJSONSchema
		}
	}
	return models.DiffFormatJSON
}

const (
	addedMarker   = "➕ added"
	changedMarker = "✏ changed"
	removedMarker = "➖ removed"

	maxTableRows  = 100
	openThreshold = 15
)

type row struct {
	path        string
	change      string
	details     string
	description string
}

// renderTable renders rows as a markdown table, capping the row count to keep comments within
// GitHub's limits. The Description column only appears when any row carries one
func renderTable(rows []row) (table string, truncated int) {
	withDescriptions := false
	for _, r := range rows {
		if r.description != "" {
			withDescriptions = true
			break
		}
	}

	var b strings.Builder
	if withDescriptions {
		b.WriteString("| Field | Change | Details | Description |\n|---|---|---|---|\n")
	} else {
		b.WriteString("| Field | Change | Details |\n|---|---|---|\n")
	}
	if len(rows) > maxTableRows {
		truncated = len(rows) - maxTableRows
		rows = rows[:maxTableRows]
	}
	for _, r := range rows {
		if withDescriptions {
			fmt.Fprintf(&b, "| `%s` | %s | %s | %s |\n",
				escapeCell(r.path), escapeCell(r.change), escapeCell(r.details), escapeCell(r.description))
		} else {
			fmt.Fprintf(&b, "| `%s` | %s | %s |\n", escapeCell(r.path), escapeCell(r.change), escapeCell(r.details))
		}
	}
	return b.String(), truncated
}

// escapeCell keeps user content from breaking out of the markdown table
func escapeCell(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
