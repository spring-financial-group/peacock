package utils

import (
	"os"
	"path/filepath"
)

func CreateTestDir(additionalPath string) (string, string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	baseDir, err := os.MkdirTemp(wd, "peacock-test")
	if err != nil {
		return "", "", err
	}
	fullPath := filepath.Join(baseDir, additionalPath)
	return baseDir, fullPath, os.MkdirAll(fullPath, 0775)
}
