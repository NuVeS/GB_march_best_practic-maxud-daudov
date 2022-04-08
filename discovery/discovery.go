package discovery

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"practic/reader"
	"practic/types"

	"go.uber.org/zap"
)

type Discovery struct {
	log    *zap.Logger
	ctx    context.Context
	reader reader.Reader
}

type MethodDto struct {
	CurDir     string
	StarterDir string
	DLimit     int
}

func newDTO(parent MethodDto, curDir string) MethodDto {
	return MethodDto{CurDir: curDir, StarterDir: parent.StarterDir, DLimit: parent.DLimit}
}

func NewDiscovery(ctx context.Context, reader reader.Reader) *Discovery {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.EncoderConfig.TimeKey = "timestamp"
	logger, err := loggerConfig.Build()
	if err != nil {
		log.Fatal(err)
	}

	d := &Discovery{log: logger, ctx: ctx, reader: reader}
	return d
}

func (d *Discovery) IsExceedLimit(dto MethodDto) (bool, int, error) {
	fileP, err := filepath.Abs(dto.CurDir)
	if err != nil {
		d.log.Error("couldnt get path from curDir", zap.String("curDir", dto.CurDir),
			zap.String("error", err.Error()))
		return false, 0, err
	}
	fileP2, err := filepath.Abs(dto.StarterDir)
	if err != nil {
		d.log.Error("couldnt get path from starterDir", zap.String("starterDir", dto.StarterDir),
			zap.String("starterDirror", err.Error()))
		return false, 0, err
	}
	depthP := strings.Split(fileP, "/")
	depthStart := strings.Split(fileP2, "/")
	if len(depthP)-len(depthStart) > dto.DLimit {
		d.log.Info("Depth limit reached", zap.Int("dLimit", dto.DLimit))
		return true, len(depthP), nil
	}

	return false, len(depthP), nil
}

func (d *Discovery) ListDirectory(dto MethodDto) ([]types.FileInfo, error) {
	d.log.Info("New direcotry check", zap.String("curDir", dto.CurDir), zap.Int("dLimit", dto.DLimit))

	select {
	case <-d.ctx.Done():
		d.log.Info("Context finished in ListDirectory")
		return nil, nil
	default:
		time.Sleep(time.Second * 3)
		var result []types.FileInfo
		res, err := d.reader.ReadDir(dto.CurDir)
		if err != nil {
			d.log.Error("couldnt read dir", zap.String("message", err.Error()))
			return nil, err
		}
		for _, entry := range res {
			path := filepath.Join(dto.CurDir, entry.Name())
			childDto := newDTO(dto, path)
			isExceed, curDepth, err := d.IsExceedLimit(childDto)
			if err != nil {
				return nil, err
			}

			if isExceed {
				d.log.Info("Depth limit reached", zap.Int("dLimit", dto.DLimit))
				return result, nil
			}
			d.NewDirCheck(entry, dto, curDepth)

		}
		return result, nil
	}
}

func (d *Discovery) NewDirCheck(entry fs.DirEntry, dto MethodDto, curDepth int) []types.FileInfo {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGUSR1)
	path := filepath.Join(dto.CurDir, entry.Name())

	select {
	case <-sigCh:
		fmt.Printf("Directory: %s, Depth: %d", dto.CurDir, curDepth)
	default:
		if entry.IsDir() {
			d.log.Info("Starting new list directory", zap.String("path", path))
			childDto := newDTO(dto, path)
			child, err := d.ListDirectory(childDto)
			if err != nil {
				d.log.Error("couldnt start new list directory ", zap.String("path", dto.StarterDir),
					zap.Int("dLimit", dto.DLimit),
					zap.String("error", err.Error()))
				return []types.FileInfo{}
			}
			if len(child) > 0 {
				d.log.Info("Got new children ", zap.String("path", path), zap.String("child", child[0].Path))
			}

			return child
		} else {
			info, err := entry.Info()
			if err != nil {
				d.log.Error("couldnt get file info ", zap.String("filename", entry.Name()),
					zap.String("error", err.Error()))
				return []types.FileInfo{}
			}
			d.log.Info("Got new children ", zap.String("path", path), zap.String("path", path))
			return []types.FileInfo{types.FileInfo{info, path}}
		}
	}
	return []types.FileInfo{}
}

func (d *Discovery) FindFiles(ext string) (types.FileList, error) {
	wd, err := d.reader.Getwd()
	if err != nil {
		d.log.Error("couldnt get wd", zap.String("error", err.Error()))
		return nil, err
	}
	dto := MethodDto{CurDir: wd, StarterDir: wd, DLimit: 2}
	files, err := d.ListDirectory(dto)
	if err != nil {
		d.log.Error("couldnt get files from dir", zap.String("wd", wd),
			zap.String("error", err.Error()))
		return nil, err
	}
	fl := make(types.FileList, len(files))
	for _, file := range files {
		if filepath.Ext(file.Name()) == ext {
			d.log.Info("Adding file", zap.String("file", file.Name()))
			fl[file.Name()] = types.TargetFile{
				Name: file.Name(),
				Path: file.Path,
			}
		}
	}
	return fl, nil
}
