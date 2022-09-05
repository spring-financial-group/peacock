package utils

import "os"

func NewPtr[T any](obj T) *T { return &obj }

// ExistsInSlice ranges over a slice returning true if the input exists in the slice
func ExistsInSlice(a string, b []string) bool {
	for _, bVal := range b {
		if bVal == a {
			return true
		}
	}
	return false
}

// Exists returns whether the given file or directory Exists
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
