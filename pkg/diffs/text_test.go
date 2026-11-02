package diffs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_textHandler_Diff(t *testing.T) {
	testCases := []struct {
		name           string
		key            string
		base           []byte
		head           []byte
		expectSections bool
		expectedSumary string
		bodyContains   []string
	}{
		{
			name:           "Changed",
			key:            "notes/config.yaml",
			base:           []byte("line1\nline2"),
			head:           []byte("line1\nline2\nline3"),
			expectSections: true,
			expectedSumary: "<code>config.yaml</code> — changed",
			bodyContains:   []string{"```diff", "+line3"},
		},
		{
			name:           "NewFile",
			key:            "notes/config.yaml",
			base:           nil,
			head:           []byte("hello"),
			expectSections: true,
			expectedSumary: "<code>config.yaml</code> — new file",
			bodyContains:   []string{"```diff", "+hello"},
		},
		{
			name:           "DeletedFile",
			key:            "notes/config.yaml",
			base:           []byte("hello"),
			head:           nil,
			expectSections: true,
			expectedSumary: "<code>config.yaml</code> — removed",
			bodyContains:   []string{"```diff", "-hello"},
		},
		{
			name:           "Unchanged",
			key:            "notes/config.yaml",
			base:           []byte("hello"),
			head:           []byte("hello"),
			expectSections: false,
		},
		{
			name:           "BothNil",
			key:            "notes/config.yaml",
			base:           nil,
			head:           nil,
			expectSections: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			sections, err := (textHandler{}).Diff(tt.key, tt.base, tt.head)
			assert.NoError(t, err)
			if !tt.expectSections {
				assert.Empty(t, sections)
				return
			}
			assert.Len(t, sections, 1)
			assert.Equal(t, tt.expectedSumary, sections[0].Summary)
			assert.True(t, sections[0].Open)
			for _, s := range tt.bodyContains {
				assert.Contains(t, sections[0].Body, s)
			}
		})
	}
}

func Test_textHandler_Diff_TruncatesLargeDiffs(t *testing.T) {
	head := ""
	for i := 0; i < maxDiffLines+100; i++ {
		head += "line\n"
	}
	sections, err := (textHandler{}).Diff("big.txt", []byte(""), []byte(head))
	assert.NoError(t, err)
	assert.Len(t, sections, 1)
	assert.Contains(t, sections[0].Body, "… diff truncated")
}
