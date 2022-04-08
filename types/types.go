package types

import (
	"os"

	"go.uber.org/zap"
)

type TargetFile struct {
	Path string
	Name string
}

type FileList map[string]TargetFile

type FileInfoType interface {
	os.FileInfo
	Path() string
}

type FileInfo struct {
	os.FileInfo
	path string
}

func (fi FileInfo) Path() string {
	return fi.path
}

type Discovery struct {
	Log *zap.Logger
}
