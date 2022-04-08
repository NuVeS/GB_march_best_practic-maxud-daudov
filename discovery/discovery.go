package discovery

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"practic/types"

	"go.uber.org/zap"
)

type Discovery struct {
	log *zap.Logger
}

func NewDiscovery() *Discovery {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.EncoderConfig.TimeKey = "timestamp"
	logger, err := loggerConfig.Build()
	if err != nil {
		log.Fatal(err)
	}

	d := &Discovery{log: logger}
	return d
}

func (d *Discovery) ListDirectory(ctx context.Context, curDir string, starterDir string, dLimit int) ([]types.FileInfo, error) {
	d.log.Info("New direcotry check", zap.String("curDir", curDir), zap.Int("dLimit", dLimit))
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGUSR1)
	select {
	case <-ctx.Done():
		d.log.Info("Context finished in ListDirectory")
		return nil, nil
	default:
		time.Sleep(time.Second * 10)
		var result []types.FileInfo
		res, err := os.ReadDir(curDir)
		if err != nil {
			d.log.Error("couldnt read dir", zap.String("message", err.Error()))
			return nil, err
		}
		for _, entry := range res {
			path := filepath.Join(curDir, entry.Name())
			fileP, err := filepath.Abs(path)
			if err != nil {
				d.log.Error("couldnt get path from curDir", zap.String("curDir", curDir),
					zap.String("error", err.Error()))
				return nil, err
			}
			fileP2, err := filepath.Abs(starterDir)
			if err != nil {
				d.log.Error("couldnt get path from starterDir", zap.String("starterDir", starterDir),
					zap.String("erstarterDirror", err.Error()))
				return nil, err
			}
			depthP := strings.Split(fileP, "\\")
			depthStart := strings.Split(fileP2, "\\")
			if len(depthP)-len(depthStart) > dLimit {
				d.log.Info("Depth limit reached", zap.Int("dLimit", dLimit))
				return result, nil
			}
			select {
			case <-sigCh:
				fmt.Printf("Директория: %s, Глубина: %d", curDir, len(depthP))
			default:
				if entry.IsDir() {
					d.log.Info("Starting new list directory", zap.String("path", path))
					child, err := d.ListDirectory(ctx, path, starterDir, dLimit) //Дополнительно: вынести в горутину
					if err != nil {
						d.log.Error("couldnt start new listdirectory ", zap.String("path", starterDir),
							zap.Int("dLimit", dLimit),
							zap.String("error", err.Error()))
						return nil, err
					}
					d.log.Info("Got new children ", zap.String("path", path), zap.String("child", child[0].Path))
					result = append(result, child...)
				} else {
					info, err := entry.Info()
					if err != nil {
						d.log.Error("couldnt get file info ", zap.String("filename", entry.Name()),
							zap.String("error", err.Error()))
						return nil, err
					}
					d.log.Info("Got new children ", zap.String("path", path), zap.String("path", path))
					result = append(result, types.FileInfo{info, path})
				}
			}

		}
		return result, nil
	}
}

func (d *Discovery) FindFiles(ctx context.Context, ext string) (types.FileList, error) {
	wd, err := os.Getwd()
	if err != nil {
		d.log.Error("couldnt get wd", zap.String("error", err.Error()))
		return nil, err
	}
	files, err := d.ListDirectory(ctx, wd, wd, 2)
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
