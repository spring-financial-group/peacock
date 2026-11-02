package diffs

import (
	"github.com/pkg/errors"
	"github.com/spring-financial-group/peacock/pkg/models"
	"gopkg.in/yaml.v3"
	"sort"
)

type configmapHandler struct{}

type configMapFile struct {
	Data map[string]string `yaml:"data"`
}

// Diff unwraps the data field of each side and diffs each key's document with its inner format.
// Only data is read — metadata (labels, annotations) is deliberately ignored so opting in stays
// entirely in the feathers config rather than in the rendered manifests.
func (configmapHandler) Diff(path string, base, head []byte) ([]models.DiffSection, error) {
	var baseData, headData map[string]string
	var err error
	if base != nil {
		baseData, err = unwrapConfigMapData(base)
		if err != nil {
			return textHandler{}.Diff(path, base, head)
		}
	}
	if head != nil {
		headData, err = unwrapConfigMapData(head)
		if err != nil {
			return textHandler{}.Diff(path, base, head)
		}
	}

	keys := make(map[string]bool)
	for k := range baseData {
		keys[k] = true
	}
	for k := range headData {
		keys[k] = true
	}
	var sorted []string
	for k := range keys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	var sections []models.DiffSection
	for _, k := range sorted {
		bv, inBase := baseData[k]
		hv, inHead := headData[k]
		if inBase && inHead && bv == hv {
			continue
		}

		innerFormat := models.DiffFormatText
		if inHead {
			innerFormat = detectInnerFormat([]byte(hv))
		} else if inBase {
			innerFormat = detectInnerFormat([]byte(bv))
		}
		inner, err := getHandler(innerFormat)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to diff configmap key %s", k)
		}
		sec, err := inner.Diff(k, valueOrNil(inBase, bv), valueOrNil(inHead, hv))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to diff configmap key %s", k)
		}
		sections = append(sections, sec...)
	}
	return sections, nil
}

// unwrapConfigMapData parses the data field out of a ConfigMap; a missing data field is not an
// error, it just leaves nothing to diff
func unwrapConfigMapData(b []byte) (map[string]string, error) {
	var f configMapFile
	if err := yaml.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	return f.Data, nil
}

func valueOrNil(present bool, s string) []byte {
	if !present {
		return nil
	}
	return []byte(s)
}
