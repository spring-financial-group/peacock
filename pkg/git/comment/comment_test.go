package comment

import "testing"

func TestGetHashFromComment(t *testing.T) {
	testCases := []struct {
		name           string
		inputComment   string
		expectedOutput string
	}{
		{
			name:           "no hash",
			inputComment:   "this is a comment",
			expectedOutput: "",
		},
		{
			name:           "with hash",
			inputComment:   "this is a comment\n <!-- hash: 1234567890 -->",
			expectedOutput: "1234567890",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := GetHashFromComment(tc.inputComment); got != tc.expectedOutput {
				t.Errorf("GetHashFromComment() = %v, want %v", got, tc.expectedOutput)
			}
		})
	}
}
