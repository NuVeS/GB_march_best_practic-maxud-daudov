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

type FileInfo struct {
	os.FileInfo
	Path string
}

type Discovery struct {
	Log *zap.Logger
}
