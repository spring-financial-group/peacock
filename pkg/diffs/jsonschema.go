package diffs

import (
	"encoding/json"
	"fmt"
	"github.com/spring-financial-group/peacock/pkg/models"
	"slices"
	"sort"
	"strings"
)

const maxRefDepth = 50

type jsonSchemaHandler struct{}

// fieldSig captures the parts of a schema field reviewers care about; descriptions are deliberately
// excluded so wording-only changes don't spam the comment
type fieldSig struct {
	types      []string
	enum       []string
	format     string
	required   bool
	deprecated bool
	ref        string
}

func (jsonSchemaHandler) Diff(key string, base, head []byte) ([]models.DiffSection, error) {
	var baseFields, headFields map[string]fieldSig
	var err error
	if base != nil {
		baseFields, err = flattenSchema(base)
		if err != nil {
			return nil, err
		}
	}
	if head != nil {
		headFields, err = flattenSchema(head)
		if err != nil {
			return nil, err
		}
	}

	switch {
	case base == nil:
		return []models.DiffSection{{
			Key:     key,
			Summary: fmt.Sprintf("<code>%s</code> — %d field%s (new)", escapeCell(key), len(headFields), plural(len(headFields))),
		}}, nil
	case head == nil:
		return []models.DiffSection{{
			Key:     key,
			Summary: fmt.Sprintf("<code>%s</code> — removed", escapeCell(key)),
		}}, nil
	}

	rows := diffFieldSigs(baseFields, headFields)
	if len(rows) == 0 {
		return nil, nil
	}
	table, truncated := renderTable(rows)
	return []models.DiffSection{{
		Key:           key,
		Summary:       fmt.Sprintf("<code>%s</code> — %d field%s changed", escapeCell(key), len(rows), plural(len(rows))),
		Body:          table,
		Open:          len(rows) <= openThreshold,
		TruncatedRows: truncated,
	}}, nil
}

// flattenSchema walks a JSON Schema document into field paths, resolving local $defs references
func flattenSchema(b []byte) (map[string]fieldSig, error) {
	var root map[string]any
	if err := json.Unmarshal(b, &root); err != nil {
		return nil, err
	}
	w := &schemaWalker{root: root, out: make(map[string]fieldSig)}
	w.walkProperties(root, "", 0)
	return w.out, nil
}

type schemaWalker struct {
	root map[string]any
	out  map[string]fieldSig
}

func (w *schemaWalker) walkProperties(node map[string]any, path string, depth int) {
	props, _ := node["properties"].(map[string]any)
	required := stringSet(node["required"])
	for name, child := range props {
		childPath := name
		if path != "" {
			childPath = path + "." + name
		}
		childNode, ok := child.(map[string]any)
		if !ok {
			continue
		}
		w.walkSchema(childNode, childPath, required[name], depth+1)
	}
}

func (w *schemaWalker) walkSchema(node map[string]any, path string, required bool, depth int) {
	// depth bound doubles as the guard against cyclic $defs references
	if depth > maxRefDepth {
		return
	}
	if ref, ok := node["$ref"].(string); ok {
		if target, resolved := w.resolveLocalRef(ref); resolved {
			w.walkSchema(target, path, required, depth+1)
			return
		}
		w.out[path] = fieldSig{ref: ref, required: required}
		return
	}
	if props, _ := node["properties"].(map[string]any); len(props) > 0 {
		w.walkProperties(node, path, depth)
		return
	}
	if t, _ := node["type"].(string); t == "array" {
		items, ok := node["items"].(map[string]any)
		if !ok {
			w.out[path+"[]"] = sigFromNode(node, required)
			return
		}
		w.walkSchema(items, path+"[]", required, depth+1)
		return
	}
	w.out[path] = sigFromNode(node, required)
}

// resolveLocalRef resolves #/... JSON pointers against the root document
func (w *schemaWalker) resolveLocalRef(ref string) (map[string]any, bool) {
	if !strings.HasPrefix(ref, "#/") {
		return nil, false
	}
	var cur any = w.root
	for _, seg := range strings.Split(strings.TrimPrefix(ref, "#/"), "/") {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[seg]
		if !ok {
			return nil, false
		}
	}
	target, ok := cur.(map[string]any)
	return target, ok
}

func sigFromNode(node map[string]any, required bool) fieldSig {
	sig := fieldSig{required: required}
	switch t := node["type"].(type) {
	case string:
		sig.types = []string{t}
	case []any:
		for _, v := range t {
			if s, ok := v.(string); ok {
				sig.types = append(sig.types, s)
			}
		}
	}
	sig.format, _ = node["format"].(string)
	if enum, ok := node["enum"].([]any); ok {
		for _, v := range enum {
			if v == nil {
				continue
			}
			sig.enum = append(sig.enum, renderValue(v))
		}
	}
	sig.deprecated, _ = node["deprecated"].(bool)
	return sig
}

func stringSet(v any) map[string]bool {
	out := make(map[string]bool)
	items, ok := v.([]any)
	if !ok {
		return out
	}
	for _, i := range items {
		if s, ok := i.(string); ok {
			out[s] = true
		}
	}
	return out
}

func diffFieldSigs(base, head map[string]fieldSig) []row {
	paths := make(map[string]bool)
	for p := range base {
		paths[p] = true
	}
	for p := range head {
		paths[p] = true
	}

	var rows []row
	for p := range paths {
		b, inBase := base[p]
		h, inHead := head[p]
		switch {
		case inBase && !inHead:
			rows = append(rows, row{path: p, change: removedMarker})
		case !inBase && inHead:
			rows = append(rows, row{path: p, change: addedMarker, details: h.render()})
		case !b.equal(h):
			rows = append(rows, row{path: p, change: changedMarker, details: renderSigChange(b, h)})
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].path < rows[j].path })
	return rows
}

func (s fieldSig) equal(o fieldSig) bool {
	return slices.Equal(s.types, o.types) &&
		slices.Equal(s.enum, o.enum) &&
		s.format == o.format &&
		s.required == o.required &&
		s.deprecated == o.deprecated &&
		s.ref == o.ref
}

// typeString renders the type portion of a field's details
func (s fieldSig) typeString() string {
	if s.ref != "" {
		return "ref: " + s.ref
	}
	var types []string
	nullable := false
	for _, t := range s.types {
		if t == "null" {
			nullable = true
			continue
		}
		types = append(types, t)
	}
	out := strings.Join(types, ", ")
	if nullable {
		out += ", nullable"
	}
	if s.format != "" {
		out += fmt.Sprintf(" (%s)", s.format)
	}
	return out
}

// render is the details shown for an added field
func (s fieldSig) render() string {
	var parts []string
	if ts := s.typeString(); ts != "" {
		parts = append(parts, ts)
	}
	if len(s.enum) > 0 {
		parts = append(parts, "enum: "+strings.Join(s.enum, ", "))
	}
	if s.deprecated {
		parts = append(parts, "deprecated")
	}
	return strings.Join(parts, "; ")
}

// renderSigChange renders each differing component as old → new
func renderSigChange(oldSig, newSig fieldSig) string {
	var parts []string
	if !slices.Equal(oldSig.types, newSig.types) || oldSig.format != newSig.format || oldSig.ref != newSig.ref {
		parts = append(parts, oldSig.typeString()+" → "+newSig.typeString())
	}
	if !slices.Equal(oldSig.enum, newSig.enum) {
		parts = append(parts, "enum: "+strings.Join(oldSig.enum, ", ")+" → enum: "+strings.Join(newSig.enum, ", "))
	}
	if oldSig.required != newSig.required {
		if newSig.required {
			parts = append(parts, "optional → required")
		} else {
			parts = append(parts, "required → optional")
		}
	}
	if oldSig.deprecated != newSig.deprecated {
		if newSig.deprecated {
			parts = append(parts, "not deprecated → deprecated")
		} else {
			parts = append(parts, "deprecated → not deprecated")
		}
	}
	return strings.Join(parts, "; ")
}
