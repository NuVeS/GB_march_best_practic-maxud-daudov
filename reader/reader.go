package reader

import (
	"io/fs"
	"os"
)

type Reader interface {
	ReadDir(name string) ([]fs.DirEntry, error) // по хорошему нужно обертку сделать над FS, но как это
	// сделать малой кровью не знаю
	Getwd() (string, error)
}

type FSReader struct {
}

func (r *FSReader) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}

func (r *FSReader) Getwd() (string, error) {
	return os.Getwd()
}
