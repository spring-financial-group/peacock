package diffs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

const (
	schemaA  = `{"$schema": "https://json-schema.org/draft/2020-12/schema", "properties": {"a": {"type": "string"}}}`
	schemaA2 = `{"$schema": "https://json-schema.org/draft/2020-12/schema", "properties": {"a": {"type": "integer"}, "b": {"type": "string"}}}`
	schemaB  = `{"$schema": "https://json-schema.org/draft/2020-12/schema", "properties": {"b": {"type": "number"}}}`
)

type cmFixture struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name        string            `yaml:"name"`
		Annotations map[string]string `yaml:"annotations"`
	} `yaml:"metadata"`
	Data map[string]string `yaml:"data"`
}

func marshalCM(t *testing.T, annotations map[string]string, data map[string]string) []byte {
	t.Helper()
	f := cmFixture{APIVersion: "v1", Kind: "ConfigMap"}
	f.Metadata.Name = "some-schema-cm"
	f.Metadata.Annotations = annotations
	f.Data = data
	b, err := yaml.Marshal(f)
	assert.NoError(t, err)
	return b
}

func Test_configmapHandler_Diff(t *testing.T) {
	t.Run("only changed data keys are rendered", func(t *testing.T) {
		base := marshalCM(t, nil, map[string]string{
			"A.json": schemaA,
			"B.json": schemaB,
		})
		head := marshalCM(t, nil, map[string]string{
			"A.json": schemaA2,
			"B.json": schemaB,
		})

		sections, err := (configmapHandler{}).Diff("config-root/ns/some-schema-cm.yaml", base, head)
		assert.NoError(t, err)
		assert.Len(t, sections, 1)
		assert.Equal(t, "A.json", sections[0].Key)
		assert.Equal(t, "<code>A.json</code> — 2 fields changed", sections[0].Summary)
		assert.Contains(t, sections[0].Body, "| `a` | ✏ changed | string → integer |")
		assert.Contains(t, sections[0].Body, "| `b` | ➕ added | string |")
	})

	t.Run("annotation only changes are omitted entirely", func(t *testing.T) {
		// schema ConfigMaps churn on every chart version bump via metadata, the data itself is what matters
		base := marshalCM(t, map[string]string{"mqube.com/extract-schema-version": "1.0.0"}, map[string]string{"A.json": schemaA})
		head := marshalCM(t, map[string]string{"mqube.com/extract-schema-version": "1.1.0"}, map[string]string{"A.json": schemaA})

		sections, err := (configmapHandler{}).Diff("config-root/ns/some-schema-cm.yaml", base, head)
		assert.NoError(t, err)
		assert.Empty(t, sections)
	})

	t.Run("new configmap renders per key counts", func(t *testing.T) {
		head := marshalCM(t, nil, map[string]string{"A.json": schemaA, "B.json": schemaB})

		sections, err := (configmapHandler{}).Diff("config-root/ns/some-schema-cm.yaml", nil, head)
		assert.NoError(t, err)
		assert.Len(t, sections, 2)
		assert.Equal(t, "<code>A.json</code> — 1 field (new)", sections[0].Summary)
		assert.Equal(t, "<code>B.json</code> — 1 field (new)", sections[1].Summary)
	})

	t.Run("removed data key renders a removed summary", func(t *testing.T) {
		base := marshalCM(t, nil, map[string]string{"A.json": schemaA, "B.json": schemaB})
		head := marshalCM(t, nil, map[string]string{"A.json": schemaA})

		sections, err := (configmapHandler{}).Diff("config-root/ns/some-schema-cm.yaml", base, head)
		assert.NoError(t, err)
		assert.Len(t, sections, 1)
		assert.Equal(t, "B.json", sections[0].Key)
		assert.Equal(t, "<code>B.json</code> — removed", sections[0].Summary)
	})

	t.Run("non json data values fall back to a text diff", func(t *testing.T) {
		base := marshalCM(t, nil, map[string]string{"notes": "hello"})
		head := marshalCM(t, nil, map[string]string{"notes": "hello world"})

		sections, err := (configmapHandler{}).Diff("config-root/ns/some-schema-cm.yaml", base, head)
		assert.NoError(t, err)
		assert.Len(t, sections, 1)
		assert.Equal(t, "notes", sections[0].Key)
		assert.Equal(t, "<code>notes</code> — changed", sections[0].Summary)
		assert.Contains(t, sections[0].Body, "+hello world")
	})

	t.Run("malformed yaml falls back to a text diff of the whole file", func(t *testing.T) {
		base := []byte("plain text, not a configmap at all")
		head := marshalCM(t, nil, map[string]string{"A.json": schemaA})

		sections, err := (configmapHandler{}).Diff("config-root/ns/some-schema-cm.yaml", base, head)
		assert.NoError(t, err)
		assert.Len(t, sections, 1)
		assert.Equal(t, "some-schema-cm.yaml", sections[0].Key)
		assert.Equal(t, "<code>some-schema-cm.yaml</code> — changed", sections[0].Summary)
	})

	t.Run("invalid inner documents surface as an error", func(t *testing.T) {
		base := marshalCM(t, nil, map[string]string{"A.json": "not json {"})
		head := marshalCM(t, nil, map[string]string{"A.json": schemaA})

		_, err := (configmapHandler{}).Diff("config-root/ns/some-schema-cm.yaml", base, head)
		assert.Error(t, err)
	})

	t.Run("binaryData only files have nothing to diff", func(t *testing.T) {
		base := []byte("apiVersion: v1\nkind: ConfigMap\nbinaryData:\n  key: aGk=\n")
		head := []byte("apiVersion: v1\nkind: ConfigMap\nbinaryData:\n  key: aGVsbG8=\n")

		sections, err := (configmapHandler{}).Diff("config-root/ns/some-schema-cm.yaml", base, head)
		assert.NoError(t, err)
		assert.Empty(t, sections)
	})

	t.Run("identical data returns no sections", func(t *testing.T) {
		base := marshalCM(t, nil, map[string]string{"A.json": schemaA})
		head := marshalCM(t, nil, map[string]string{"A.json": schemaA})

		sections, err := (configmapHandler{}).Diff("config-root/ns/some-schema-cm.yaml", base, head)
		assert.NoError(t, err)
		assert.Empty(t, sections)
	})
}
