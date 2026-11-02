package diffs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	baseSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "applicationId": {"type": "string"},
    "applicants": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": {"type": "string"},
          "firstTimeBuyer": {"type": ["boolean", "null"]}
        }
      }
    },
    "product": {
      "type": "object",
      "required": ["rate"],
      "properties": {
        "rate": {"type": "number"},
        "term": {"$ref": "#/$defs/termDto"}
      }
    }
  },
  "$defs": {
    "termDto": {
      "type": "object",
      "required": ["months"],
      "properties": {
        "months": {"type": "integer"},
        "erc": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "rate": {"type": "number"},
              "year": {"type": "integer"}
            }
          }
        }
      }
    }
  }
}`

	headSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "applicationId": {"type": "string", "description": "the application identifier"},
    "adviceType": {"type": "string", "enum": ["Independent", "NonAdvice", null]},
    "applicants": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": {"type": "string"},
          "firstTimeBuyer": {"type": ["boolean", "null"]},
          "applicationRole": {"type": "string", "enum": ["Borrower", "Guarantor", "Director"]}
        }
      }
    },
    "product": {
      "type": "object",
      "required": ["rate"],
      "properties": {
        "rate": {"type": "string"},
        "term": {"$ref": "#/$defs/termDto"}
      }
    }
  },
  "$defs": {
    "termDto": {
      "type": "object",
      "required": [],
      "properties": {
        "months": {"type": "integer"},
        "erc": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "rate": {"type": ["number", "null"]}
            }
          }
        }
      }
    }
  }
}`

	expectedSchemaTable = `| Field | Change | Details |
|---|---|---|
| ` + "`adviceType`" + ` | ➕ added | string; enum: Independent, NonAdvice |
| ` + "`applicants[].applicationRole`" + ` | ➕ added | string; enum: Borrower, Guarantor, Director |
| ` + "`product.rate`" + ` | ✏ changed | number → string |
| ` + "`product.term.erc[].rate`" + ` | ✏ changed | number → number, nullable |
| ` + "`product.term.erc[].year`" + ` | ➖ removed |  |
| ` + "`product.term.months`" + ` | ✏ changed | required → optional |
`
)

func Test_jsonSchemaHandler_Diff(t *testing.T) {
	t.Run("field level diff across refs, arrays and enums", func(t *testing.T) {
		sections, err := (jsonSchemaHandler{}).Diff("OrigoExtractEntry.json", []byte(baseSchema), []byte(headSchema))
		assert.NoError(t, err)
		assert.Len(t, sections, 1)

		s := sections[0]
		assert.Equal(t, "OrigoExtractEntry.json", s.Key)
		assert.Equal(t, "<code>OrigoExtractEntry.json</code> — 6 fields changed", s.Summary)
		assert.Equal(t, expectedSchemaTable, s.Body)
		assert.True(t, s.Open)
		assert.Equal(t, 0, s.TruncatedRows)
	})

	t.Run("identical schemas return no sections", func(t *testing.T) {
		sections, err := (jsonSchemaHandler{}).Diff("schema.json", []byte(baseSchema), []byte(baseSchema))
		assert.NoError(t, err)
		assert.Empty(t, sections)
	})

	t.Run("new schema returns a field count", func(t *testing.T) {
		sections, err := (jsonSchemaHandler{}).Diff("schema.json", nil, []byte(headSchema))
		assert.NoError(t, err)
		assert.Len(t, sections, 1)
		assert.Equal(t, "<code>schema.json</code> — 8 fields (new)", sections[0].Summary)
		assert.Empty(t, sections[0].Body)
	})

	t.Run("deleted schema returns a removed summary", func(t *testing.T) {
		sections, err := (jsonSchemaHandler{}).Diff("schema.json", []byte(baseSchema), nil)
		assert.NoError(t, err)
		assert.Len(t, sections, 1)
		assert.Equal(t, "<code>schema.json</code> — removed", sections[0].Summary)
	})

	t.Run("invalid json errors", func(t *testing.T) {
		_, err := (jsonSchemaHandler{}).Diff("schema.json", []byte("not json"), []byte(headSchema))
		assert.Error(t, err)
	})

	t.Run("added nullable field renders the nullable suffix", func(t *testing.T) {
		base := []byte(`{"$schema": "https://json-schema.org/draft/2020-12/schema", "properties": {}}`)
		head := []byte(`{"$schema": "https://json-schema.org/draft/2020-12/schema", "properties": {"firstTimeBuyer": {"type": ["boolean", "null"]}}}`)

		sections, err := (jsonSchemaHandler{}).Diff("schema.json", base, head)
		assert.NoError(t, err)
		assert.Len(t, sections, 1)
		assert.Contains(t, sections[0].Body, "| `firstTimeBuyer` | ➕ added | boolean, nullable |")
	})

	t.Run("format changes are rendered on the type", func(t *testing.T) {
		base := []byte(`{"$schema": "https://json-schema.org/draft/2020-12/schema", "properties": {"when": {"type": "string", "format": "date"}}}`)
		head := []byte(`{"$schema": "https://json-schema.org/draft/2020-12/schema", "properties": {"when": {"type": "string", "format": "date-time"}}}`)

		sections, err := (jsonSchemaHandler{}).Diff("schema.json", base, head)
		assert.NoError(t, err)
		assert.Len(t, sections, 1)
		assert.Contains(t, sections[0].Body, "| `when` | ✏ changed | string (date) → string (date-time) |")
	})

	t.Run("description only changes return no sections", func(t *testing.T) {
		base := []byte(`{"$schema": "https://json-schema.org/draft/2020-12/schema", "properties": {"a": {"type": "string"}}}`)
		head := []byte(`{"$schema": "https://json-schema.org/draft/2020-12/schema", "properties": {"a": {"type": "string", "description": "new words"}}}`)

		sections, err := (jsonSchemaHandler{}).Diff("schema.json", base, head)
		assert.NoError(t, err)
		assert.Empty(t, sections)
	})

	t.Run("descriptions render in their own column when present", func(t *testing.T) {
		base := []byte(`{"$schema": "https://json-schema.org/draft/2020-12/schema", "properties": {"gone": {"type": "string", "description": "The gone field"}}}`)
		head := []byte(`{"$schema": "https://json-schema.org/draft/2020-12/schema", "properties": {"added": {"type": "integer", "description": "The added count"}, "changed": {"type": "string", "description": "The changed value"}}}`)

		sections, err := (jsonSchemaHandler{}).Diff("schema.json", base, head)
		assert.NoError(t, err)
		assert.Len(t, sections, 1)
		assert.Contains(t, sections[0].Body, "| Field | Change | Details | Description |\n|---|---|---|---|\n")
		assert.Contains(t, sections[0].Body, "| `added` | ➕ added | integer | The added count |")
		assert.Contains(t, sections[0].Body, "| `changed` | ➕ added | string | The changed value |")
		assert.Contains(t, sections[0].Body, "| `gone` | ➖ removed |  | The gone field |")
	})

	t.Run("long descriptions are truncated", func(t *testing.T) {
		long := strings.Repeat("word ", 40)
		base := []byte(`{"$schema": "https://json-schema.org/draft/2020-12/schema", "properties": {}}`)
		head := []byte(`{"$schema": "https://json-schema.org/draft/2020-12/schema", "properties": {"a": {"type": "string", "description": "` + long + `"}}}`)

		sections, err := (jsonSchemaHandler{}).Diff("schema.json", base, head)
		assert.NoError(t, err)
		assert.Len(t, sections, 1)
		assert.Contains(t, sections[0].Body, "… |")
	})

	t.Run("cyclic refs terminate instead of hanging", func(t *testing.T) {
		cyclic := []byte(`{"$schema": "https://json-schema.org/draft/2020-12/schema", "properties": {"loop": {"$ref": "#/$defs/a"}}, "$defs": {"a": {"$ref": "#/$defs/a"}}}`)

		sections, err := (jsonSchemaHandler{}).Diff("schema.json", cyclic, cyclic)
		assert.NoError(t, err)
		assert.Empty(t, sections)
	})

	t.Run("non local refs become leaves", func(t *testing.T) {
		base := []byte(`{"$schema": "https://json-schema.org/draft/2020-12/schema", "properties": {"ext": {"$ref": "https://example.com/schema.json"}}}`)
		head := []byte(`{"$schema": "https://json-schema.org/draft/2020-12/schema", "properties": {"ext": {"$ref": "https://example.com/v2.json"}}}`)

		sections, err := (jsonSchemaHandler{}).Diff("schema.json", base, head)
		assert.NoError(t, err)
		assert.Len(t, sections, 1)
		assert.Contains(t, sections[0].Body, "| `ext` | ✏ changed | ref: https://example.com/schema.json → ref: https://example.com/v2.json |")
	})
}
