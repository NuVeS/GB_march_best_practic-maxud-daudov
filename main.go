package main

//Исходники задания для первого занятия у других групп https://github.com/t0pep0/GB_best_go

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
)

type TargetFile struct {
	Path string
	Name string
}

type FileList map[string]TargetFile

type FileInfo interface {
	os.FileInfo
	Path() string
}

type fileInfo struct {
	os.FileInfo
	path string
}

func (fi fileInfo) Path() string {
	return fi.path
}

//Ограничить глубину поиска заданым числом, по SIGUSR2 увеличить глубину поиска на +2
func ListDirectory(ctx context.Context, curDir string, starterDir string, dLimit int) ([]FileInfo, error) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGUSR1)
	select {
	case <-ctx.Done():
		return nil, nil
	default:
		//По SIGUSR1 вывести текущую директорию и текущую глубину поиска
		time.Sleep(time.Second * 10)
		var result []FileInfo
		res, err := os.ReadDir(curDir)
		if err != nil {
			return nil, err
		}
		for _, entry := range res {
			path := filepath.Join(curDir, entry.Name())
			fileP, err := filepath.Abs(path)
			if err != nil {
				return nil, err
			}
			fileP2, err := filepath.Abs(starterDir)
			if err != nil {
				return nil, err
			}
			depthP := strings.Split(fileP, "\\")
			depthStart := strings.Split(fileP2, "\\")
			if len(depthP)-len(depthStart) > dLimit {
				return result, nil
			}
			select {
			case <-sigCh:
				fmt.Printf("Директория: %s, Глубина: %d", curDir, len(depthP))
			default:
				if entry.IsDir() {
					listCh := make(chan error, 1)
					go func() {
						child, err := ListDirectory(ctx, path, starterDir, dLimit) //Дополнительно: вынести в горутину
						if err != nil {
							listCh <- err
							return
						}
						result = append(result, child...)
						listCh <- nil
					}()
					select {
					case err := <-listCh:
						if err != nil {
							return nil, err
						}
					}

				} else {
					info, err := entry.Info()
					if err != nil {
						return nil, err
					}
					result = append(result, fileInfo{info, path})
				}
			}

		}
		return result, nil
	}
}

func FindFiles(ctx context.Context, ext string) (FileList, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	files, err := ListDirectory(ctx, wd, wd, 2)
	if err != nil {
		return nil, err
	}
	fl := make(FileList, len(files))
	for _, file := range files {
		if filepath.Ext(file.Name()) == ext {
			fl[file.Name()] = TargetFile{
				Name: file.Name(),
				Path: file.Path(),
			}
		}
	}
	return fl, nil
}

func main() {
	const wantExt = ".go"
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	//Обработать сигнал SIGUSR1
	waitCh := make(chan struct{})
	go func() {
		res, err := FindFiles(ctx, wantExt)
		if err != nil {
			log.Printf("Error on search: %v\n", err)
			os.Exit(1)
		}
		for _, f := range res {
			fmt.Printf("\tName: %s\t\t Path: %s\n", f.Name, f.Path)
		}
		waitCh <- struct{}{}
	}()
	go func() {
		<-sigCh
		log.Println("Signal received, terminate...")
		cancel()
	}()
	//Дополнительно: Ожидание всех горутин перед завершением
	<-waitCh
	log.Println("Done")
}
