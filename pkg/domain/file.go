package domain

type FileUseCase interface {
	Read(fileName string) ([]byte, error)
	Exists(path string) (bool, error)
	Write(data []byte, path string) error
	Remove(path string) error
	Mkdir(path string) error
}
