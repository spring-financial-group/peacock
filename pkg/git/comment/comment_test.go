package comment

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetHashFromComment(t *testing.T) {
	testCases := []struct {
		name                string
		inputComment        string
		expectedHash        string
		expectedCommentType string
	}{
		{
			name:                "no hash",
			inputComment:        "this is a comment",
			expectedHash:        "",
			expectedCommentType: "",
		},
		{
			name:                "with hash and type",
			inputComment:        "this is a comment\n <!-- hash: 1234567890 type: breakdown -->",
			expectedHash:        "1234567890",
			expectedCommentType: "breakdown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash, commentType := GetMetadataFromComment(tc.inputComment)
			assert.Equal(t, tc.expectedHash, hash)
			assert.Equal(t, tc.expectedCommentType, commentType)
		})
	}
}

func TestAddMetadataToComment(t *testing.T) {
	testCases := []struct {
		name             string
		inputComment     string
		inputHash        string
		inputCommentType string
		expectedComment  string
	}{
		{
			name:             "with hash and type",
			inputComment:     "this is a comment",
			inputHash:        "1234567890",
			inputCommentType: "breakdown",
			expectedComment:  "this is a comment\n<!-- hash: 1234567890 type: breakdown -->\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			comment := AddMetadataToComment(tc.inputComment, tc.inputHash, tc.inputCommentType)
			assert.Equal(t, tc.expectedComment, comment)
		})
	}
}
