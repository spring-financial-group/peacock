package file

import (
	"github.com/spring-financial-group/peacock/pkg/utils"
	"os"
	"path/filepath"
)

type UseCase struct{}

func NewFileUseCase() UseCase {
	return UseCase{}
}

func (f UseCase) Read(fileName string) ([]byte, error) {
	return os.ReadFile(fileName)
}

func (f UseCase) Exists(path string) (bool, error) {
	return utils.Exists(path)
}

func (f UseCase) Write(data []byte, path string) error {
	err := os.MkdirAll(filepath.Dir(path), 0775)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0775)
}

func (f UseCase) Remove(path string) error {
	return os.Remove(path)
}

func (f UseCase) Mkdir(path string) error {
	return os.MkdirAll(path, 0775)
}
