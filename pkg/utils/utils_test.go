package utils_test

import (
	"github.com/spring-financial-group/peacock/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUtils_CommaSeperated(t *testing.T) {
	testCases := []struct {
		name           string
		input          []any
		expectedOutput string
	}{
		{
			name:           "strings",
			input:          []any{"one", "two", "three", "four"},
			expectedOutput: "one, two, three, four",
		},
		{
			name:           "integers",
			input:          []any{1, 2, 3, 4},
			expectedOutput: "1, 2, 3, 4",
		},
		{
			name:           "floats",
			input:          []any{0.1, 0.2, 0.3, 0.4},
			expectedOutput: "0.1, 0.2, 0.3, 0.4",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			actualOutput := utils.CommaSeperated(tt.input)
			assert.Equal(t, tt.expectedOutput, actualOutput)
		})
	}
}
