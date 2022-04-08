package reader

import (
	"io/fs"
	"os"
)

type FSReader struct {
}

func (r *FSReader) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}

func (r *FSReader) Getwd() (string, error) {
	return os.Getwd()
}
