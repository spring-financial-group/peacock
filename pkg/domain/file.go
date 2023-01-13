package domain

import "fmt"

type FileUseCase interface {
	Read(fileName string) ([]byte, error)
	Exists(path string) (bool, error)
	Write(data []byte, path string) error
	Remove(path string) error
	Mkdir(path string) error
}

type ErrFileNotFound struct {
	Path string
}

func (e *ErrFileNotFound) Error() string {
	return fmt.Sprintf("file not found: %s", e.Path)
}
