package diffs

import (
	"encoding/json"
	"fmt"
	"github.com/spring-financial-group/peacock/pkg/models"
	"sort"
)

const maxValueLen = 80

type jsonHandler struct{}

func (jsonHandler) Diff(key string, base, head []byte) ([]models.DiffSection, error) {
	var basePaths, headPaths map[string]string
	var err error
	if base != nil {
		basePaths, err = flattenJSON(base)
		if err != nil {
			return nil, err
		}
	}
	if head != nil {
		headPaths, err = flattenJSON(head)
		if err != nil {
			return nil, err
		}
	}

	switch {
	case base == nil:
		return []models.DiffSection{{
			Key:     key,
			Summary: fmt.Sprintf("<code>%s</code> — %d path%s (new)", escapeCell(key), len(headPaths), plural(len(headPaths))),
		}}, nil
	case head == nil:
		return []models.DiffSection{{
			Key:     key,
			Summary: fmt.Sprintf("<code>%s</code> — removed", escapeCell(key)),
		}}, nil
	}

	var rows []row
	for _, p := range sortedKeys(basePaths, headPaths) {
		bv, inBase := basePaths[p]
		hv, inHead := headPaths[p]
		switch {
		case inBase && !inHead:
			rows = append(rows, row{path: p, change: removedMarker})
		case !inBase && inHead:
			rows = append(rows, row{path: p, change: addedMarker, details: hv})
		case bv != hv:
			rows = append(rows, row{path: p, change: changedMarker, details: bv + " → " + hv})
		}
	}
	if len(rows) == 0 {
		return nil, nil
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].path < rows[j].path })

	table, truncated := renderTable(rows)
	return []models.DiffSection{{
		Key:           key,
		Summary:       fmt.Sprintf("<code>%s</code> — %d path%s changed", escapeCell(key), len(rows), plural(len(rows))),
		Body:          table,
		Open:          len(rows) <= openThreshold,
		TruncatedRows: truncated,
	}}, nil
}

// flattenJSON flattens any JSON document to dot-separated paths with array indices
func flattenJSON(b []byte) (map[string]string, error) {
	var doc any
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, err
	}
	out := make(map[string]string)
	flattenValue(doc, "", out)
	return out, nil
}

func flattenValue(node any, prefix string, out map[string]string) {
	switch v := node.(type) {
	case map[string]any:
		for k, child := range v {
			p := k
			if prefix != "" {
				p = prefix + "." + k
			}
			flattenValue(child, p, out)
		}
	case []any:
		for i, child := range v {
			flattenValue(child, fmt.Sprintf("%s[%d]", prefix, i), out)
		}
	default:
		out[prefix] = truncate(renderValue(node))
	}
}

// renderValue keeps strings readable and encodes everything else as compact JSON
func renderValue(v any) string {
	if v == nil {
		return "null"
	}
	if s, ok := v.(string); ok {
		return s
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

func truncate(s string) string {
	if len(s) <= maxValueLen {
		return s
	}
	return s[:maxValueLen-1] + "…"
}

func sortedKeys(a, b map[string]string) []string {
	seen := make(map[string]bool)
	var keys []string
	for k := range a {
		seen[k] = true
		keys = append(keys, k)
	}
	for k := range b {
		if !seen[k] {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}
