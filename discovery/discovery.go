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

type methodDto struct {
	curDir     string
	starterDir string
	dLimit     int
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

func (d *Discovery) isExceedLimit(dto methodDto) (bool, int, error) {
	fileP, err := filepath.Abs(dto.curDir)
	if err != nil {
		d.log.Error("couldnt get path from curDir", zap.String("curDir", dto.curDir),
			zap.String("error", err.Error()))
		return false, 0, err
	}
	fileP2, err := filepath.Abs(dto.starterDir)
	if err != nil {
		d.log.Error("couldnt get path from starterDir", zap.String("starterDir", dto.starterDir),
			zap.String("starterDirror", err.Error()))
		return false, 0, err
	}
	depthP := strings.Split(fileP, "\\")
	depthStart := strings.Split(fileP2, "\\")
	if len(depthP)-len(depthStart) > dto.dLimit {
		d.log.Info("Depth limit reached", zap.Int("dLimit", dto.dLimit))
		return true, len(depthP), nil
	}

	return false, len(depthP), nil
}

func (d *Discovery) ListDirectory(dto methodDto) ([]types.FileInfo, error) {
	d.log.Info("New direcotry check", zap.String("curDir", dto.curDir), zap.Int("dLimit", dto.dLimit))

	select {
	case <-d.ctx.Done():
		d.log.Info("Context finished in ListDirectory")
		return nil, nil
	default:
		time.Sleep(time.Second * 10)
		var result []types.FileInfo
		res, err := d.reader.ReadDir(dto.curDir)
		if err != nil {
			d.log.Error("couldnt read dir", zap.String("message", err.Error()))
			return nil, err
		}
		for _, entry := range res {
			path := filepath.Join(dto.curDir, entry.Name())
			childDto := methodDto{curDir: path, starterDir: dto.starterDir, dLimit: dto.dLimit}
			isExceed, curDepth, err := d.isExceedLimit(childDto)
			if err != nil {
				return nil, err
			}

			if isExceed {
				d.log.Info("Depth limit reached", zap.Int("dLimit", dto.dLimit))
				return result, nil
			}
			d.newDirCheck(entry, dto, curDepth)

		}
		return result, nil
	}
}

func (d *Discovery) newDirCheck(entry fs.DirEntry, dto methodDto, curDepth int) []types.FileInfo {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGUSR1)
	path := filepath.Join(dto.curDir, entry.Name())

	select {
	case <-sigCh:
		fmt.Printf("Directory: %s, Depth: %d", dto.curDir, curDepth)
	default:
		if entry.IsDir() {
			d.log.Info("Starting new list directory", zap.String("path", path))
			childDto := methodDto{curDir: path, starterDir: dto.starterDir, dLimit: dto.dLimit}
			child, err := d.ListDirectory(childDto)
			if err != nil {
				d.log.Error("couldnt start new list directory ", zap.String("path", dto.starterDir),
					zap.Int("dLimit", dto.dLimit),
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
	dto := methodDto{curDir: wd, starterDir: wd, dLimit: 2}
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
